package dal

import (
	"demo_thrift/biz/dal/mysql"
	"demo_thrift/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
