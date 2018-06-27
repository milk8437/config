package config

import (
	"os"
	"io/ioutil"
	"github.com/BurntSushi/toml"
	"fmt"
)

//加载配置文件的方法
type LoadDataFromEsc func(env ENV, useLocal bool) (string)

var Prop Properties

type Properties struct {
	Server   *Server
	Logger   *Logger
	Database *Database
	Redis    *RedisConfig
	Kafka    *KafkaConfig
	Variable map[string]string
}

//加载解析配置文件
func NewProperties(ld LoadDataFromEsc) {
	command := parseCommand()
	var data string

	if command.Local && len(command.Fpath) > 0 {
		fmt.Println("load config file:" + command.Fpath)
		//加载命令行指定配置文件
		file, err := os.Open(command.Fpath)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		data = string(bytes)
	} else {
		data = ld(command.Env, command.Local)
	}

	_, err := toml.Decode(data, &Prop)
	if err != nil {
		panic(err)
	}

	if command.Port != -1 {
		Prop.Server.Port = command.Port
	}
}
