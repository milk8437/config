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
func CreateRedis(config map[string]interface{}) (func(), error) {
    rc :=Prop.Database
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
	return nil, nil

}

type RedisConfig struct {
	Port     int
	Host     string
	Password string
}
