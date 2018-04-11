package redispack

import (
	"github.com/garyburd/redigo/redis"
	"gylib/common"
	"time"
)

func Get_redis_pool() *redis.Pool {
	data := common.Getini("conf/app.ini","redis",map[string]string{"redis_host":"127.0.0.1","redis_port":"6379","redis_auth":"","redis_db":"0"})
	return &redis.Pool{
		MaxIdle:     100,
		MaxActive:   1000,
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
