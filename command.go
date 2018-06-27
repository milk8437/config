package config

import "flag"

type ENV string

const (
	UAT   ENV = "uat"
	PEA   ENV = "peA"
	PEB   ENV = "peB"
	LOCAL ENV = "local"
)

//CommandConfig 命令行配置信息
//Local 是否使用本地配置文件(local)
//Port server的端口(port)
//Env 使用的配置文件(env)
type commandConfig struct {
	Env   ENV
	Port  int
	Local bool
	Fpath string
}

//parseCommand 解析命令行
//port 默认值为-1 服务端口
func parseCommand() *commandConfig {
	env := flag.String("env", "uat", "env config file...")
	port := flag.Int("port", -1, "server port >=3000,default 9090")
	local := flag.Bool("local", true, "use local config  file or memory file,default true")
	fpath := flag.String("fpath", "", "config file ")
	flag.Parse()
	return &commandConfig{
		ENV(*env),
		*port,
		*local,
		*fpath,
	}
}

func (cc *commandConfig) serverPort() (int, bool) {
	if cc.Port == -1 {
		return cc.Port, false
	}
	return cc.Port, true
}
