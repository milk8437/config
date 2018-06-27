package config

import (
	"time"
	"go.uber.org/zap"
	"github.com/jackc/pgx"
)

var PGClient *pgx.ConnPool

//数据库配置
type Database struct {
	Username string
	Password string
	Host     string
	Port     int
	DBname   string `toml:"dbname"`
}

// 实现 context parse func
func CreateDB() {
	dc := Prop.Database
	if dc.Port == 0 {
		dc.Port = 5432
	}
	poolConfig := pgx.ConnPoolConfig{
		MaxConnections: 20,
		ConnConfig: pgx.ConnConfig{
			Host:     dc.Host,
			Port:     uint16(dc.Port),
			Database: dc.DBname,
			User:     dc.Username, // default: OS user name
			Password: dc.Password,
		},
		AcquireTimeout: time.Second * 30,
	}

	var err error
	if PGClient, err = pgx.NewConnPool(poolConfig); err != nil {
		panic(err)
	}
	log := LOG.Named("POSTGRES")
	log.Info("create POSTGRES client successfully...",
		zap.String("host", dc.Host))
}
