package dal

import (
	"demo_proto/biz/dal/mysql"
	"demo_proto/biz/dal/redis"
)

func Init() {
	redis.Init()
	mysql.Init()
}
