package kv

import (
	"log"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/oldfritter/lucy/base"
)

var (
	dataPool *redis.Pool
)

func init() {
	if dataPool == nil {
		dataPool = newRedisPool("data")
	}
}

func CloseRedisPools() {
	dataPool.Close()
}

func GetRedisConn(redisName string) redis.Conn {
	return dataPool.Get()
}

func newRedisPool(redisName string) *redis.Pool {
	redisConfig := base.GetConfig()
	capacity := redisConfig.GetInt("redis."+redisName+".pool", 10)
	maxCapacity := redisConfig.GetInt("redis."+redisName+".maxopen", 0)
	idleTimout := redisConfig.GetDuration("redis."+redisName+".timeout", "4m")
	maxConnLifetime := redisConfig.GetDuration("redis."+redisName+".life_time", "2m")
	network := redisConfig.Get("redis."+redisName+".network", "tcp")
	server := redisConfig.Get("redis."+redisName+".server", "localhost:6379")
	db := redisConfig.Get("redis."+redisName+".db", "")
	username := redisConfig.Get("redis."+redisName+".username", "")
	password := redisConfig.Get("redis."+redisName+".password", "")

	return &redis.Pool{
		MaxIdle:         capacity,
		MaxActive:       maxCapacity,
		IdleTimeout:     idleTimout,
		MaxConnLifetime: maxConnLifetime,
		Wait:            true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(network, server)
			if err != nil {
				log.Println("redis can't dial:" + err.Error())
				return nil, err
			}

			if username != "" && password != "" {
				if _, err := conn.Do("AUTH", username, password); err != nil {
					log.Println("redis can't AUTH:" + err.Error())
					conn.Close()
					return nil, err
				}
			}

			if db != "" {
				_, err := conn.Do("SELECT", db)
				if err != nil {
					log.Println("redis can't SELECT:" + err.Error())
					conn.Close()
					return nil, err
				}
			}
			log.Printf("redis %s connect success!", redisName)
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				log.Println("redis can't ping, err:" + err.Error())
			}
			return err
		},
	}
}
