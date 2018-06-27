package config

import (
	"strconv"
	"go.uber.org/zap"
	"github.com/go-redis/redis"
)

var (
	RedisClient *redis.Client
)

//构造redis客户端
func CreateRedis() {
    rc :=Prop.Redis
	if rc.Port == 0 {
		rc.Port = 6379
	}
	RedisClient = redis.NewClient(&redis.Options{
		Password: rc.Password,
		Addr:     rc.Host + ":" + strconv.Itoa(rc.Port),
	})
	log := LOG.Named("REDIS")
	log.Info("create REDIS client successfully...",
		zap.String("host", rc.Host))
}

type RedisConfig struct {
	Port     int
	Host     string
	Password string
}
