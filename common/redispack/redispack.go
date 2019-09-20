package redispack

import (
	"github.com/garyburd/redigo/redis"
	"gylib/common"
	"time"
	"gylib/common/datatype"
)

var Redis_data map[string]string

func init() {
	Redis_data = make(map[string]string)
	Redis_data = common.Getini("conf/app.ini", "redis", map[string]string{"redis_host": "127.0.0.1", "redis_port": "6379", "redis_auth": "", "redis_db": "0", "redis_perfix": "", "redis_minpool": "5", "redis_maxpool": "20"})
}

func Get_redis_pool() *redis.Pool {
	data := Redis_data
	return &redis.Pool{
		MaxIdle:     datatype.Str2Int(Redis_data["redis_minpool"]),
		MaxActive:   datatype.Str2Int(Redis_data["redis_maxpool"]),
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", data["redis_host"]+":"+data["redis_port"])
			if err != nil {
				return nil, err
			}
			// 选择db
			if (data["redis_auth"] != "") {
				c.Do("AUTH", data["redis_auth"])
			}
			c.Do("SELECT", data["redis_db"])

			return c, nil
		},
	}
}
