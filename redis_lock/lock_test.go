package redis_lock

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func Test_blockingLock(t *testing.T) {
	addr := "127.0.0.1:6378"
	passwd := ""

	client := NewClient("tcp", addr, passwd)
	lock1 := NewRedisLock("test1_key", client, WithExpireSeconds(1))
	lock2 := NewRedisLock("test1_key", client, WithBlock(), WithExpireSeconds(2))


	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock2.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Wait()

	t.Log("success")
}

func Test_nonblockingLock(t *testing.T) {
	addr := "127.0.0.1:6378"
	passwd := ""

	client := NewClient("tcp", addr, passwd)
	lock1 := NewRedisLock("test2_key", client, WithExpireSeconds(1))
	lock2 := NewRedisLock("test2_key", client)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock2.Lock(ctx); err == nil || !errors.Is(err, ErrLockAcquiredByOthers) {
			t.Errorf("got err: %v, expect: %v", err, ErrLockAcquiredByOthers)
			return
		}
	}()

	wg.Wait()
	t.Log("success")
}