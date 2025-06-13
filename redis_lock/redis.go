package redis_lock

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

type LockClient interface {
	SetNX(ctx context.Context, key, value string, expireSeconds int64) (int64, error)
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
}

// Client Redis 客户端.
type Client struct {
	ClientOptions
	pool *redis.Pool
}

func NewClient(network, address, password string, options ...ClientOption) *Client {

	c := Client{
		ClientOptions: ClientOptions{
			network:  network,
			address:  address,
			password: password,
		},
	}

	for _, option := range options {
		option(&c.ClientOptions)
	}

	repairClient(&c.ClientOptions)

	pool := c.getRedisPool()
	return &Client{pool: pool}
}

func (c *Client) getRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     c.maxIdle,
		IdleTimeout: time.Duration(c.idleTimeoutSeconds) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := c.getRedisConn()
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxActive: c.maxActive,
		Wait:      c.wait,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("ping")
			return err
		},
	}
}

func (c *Client) getRedisConn() (redis.Conn, error) {
	if c.address == "" {
		panic("Cannot get redis address from config")
	}

	var dialOpts []redis.DialOption
	if len(c.password) > 0 {
		dialOpts = append(dialOpts, redis.DialPassword(c.password))
	}
	conn, err := redis.DialContext(context.Background(), c.network, c.address, dialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Client) GetConn(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// 只有 key 不存在时，能够 set 成功. set 时携带上超时时间，单位秒.
func (c *Client) SetNEX(ctx context.Context, key, value string, expireSeconds int64) (int64, error) {
	if key == "" || value == "" {
		return -1, errors.New("redis SET keyNX or value can't be empty")
	}

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	reply, err := conn.Do("SET", key, value, "EX", expireSeconds, "NX")
	if err != nil {
		return -1, err
	}

	r, _ := reply.(int64)
	return r, nil
}

// Eval 支持使用 lua 脚本.
func (c *Client) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
	args := make([]interface{}, 0, 2*len(keysAndArgs))
	args[0] = src
	args[1] = keyCount
	copy(args[2:], keysAndArgs)

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	return conn.Do("EVAL", args...)
}
