package config

import (
	"github.com/spf13/viper"
	"tunnel/pkg/log"
)

type Config struct {
	Http struct {
		IP   string
		Port int
		Mode string
	}
	DependOnServices struct {
		User struct {
			Address string
		}
	}
	Mysql struct {
		DSN         string
		MaxLifeTime int
		MaxOpenConn int
		MaxIdleConn int
	}
	Redis struct {
		Host string
		Port int
		Pwd  string `mapstructure:"pwd"`
	}
	Log struct {
		Level   string
		LogPath string `mapstructure:"logPath"`
	}
	TunnelServer struct {
		IP         string
		MinPort    int
		MaxPort    int
		Image      string
		AppName    string
		ConfigName string `mapstructure:"configName"`
		Replicas   int
		Namespace  string
	}
	TunnelClient struct {
		AppName    string
		Image      string
		ConfigName string `mapstructure:"configName"`
		RootDomain string `mapstructure:"rootDomain"`
	}
	AliYunDomain struct {
		AccessKeyID     string `mapstructure:"accessKeyId"`
		AccessKeySecret string `mapstructure:"accessKeySecret"`
		Endpoint        string `mapstructure:"endpoint"`
		RootDomain      string `mapstructure:"rootDomain"`
	} `mapstructure:"aliYunDomain"`
	NginxGateway struct {
		Image      string
		AppName    string
		ConfigPath string
		Port       uint32
		Namespace  string
	}
}

var conf *Config

func InitConfig(filePath string, typ ...string) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if len(typ) > 0 {
		v.SetConfigType(typ[0])
	}
	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	conf = &Config{}
	err = v.Unmarshal(conf)
	if err != nil {
		log.Fatal(err)
	}
}
func GetConfig() *Config {
	return conf
}
