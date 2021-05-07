package config

import (
	"log"
	"sync/atomic"
)

type AppConfig struct {
	ListenServer string
	ListenPath   string
	IsAutoReLoad bool
}

type AppConfigMgr struct {
	Config atomic.Value
}

var AppConfigManager = &AppConfigMgr{}

func (a *AppConfigMgr) Callback(conf *Config) {
	appConfig := &AppConfig{}
	listenServer, err := conf.GetString("listenServer")
	if err != nil {
		log.Printf("get listenServer err: %v\n", err)
		return
	}
	appConfig.ListenServer = listenServer

	listenPath, err := conf.GetString("listenPath")
	if err != nil {
		log.Printf("get listenPath err: %v\n", err)
		return
	}
	appConfig.ListenPath = listenPath

	AppConfigManager.Config.Store(appConfig)
}

func InitConfig(file string) {
	conf, err := NewConfig(file)
	if err != nil {
		log.Printf("read config file err: %v\n", err)
		return
	}

	conf.AddObserver(AppConfigManager)

	var appConfig AppConfig
	appConfig.ListenServer, err = conf.GetString("listenServer")
	if err != nil {
		log.Printf("get listenServer err: %v\n", err)
		return
	}
	log.Println("listenServer: ", appConfig.ListenServer)

	appConfig.ListenPath, err = conf.GetString("listenPath")
	if err != nil {
		log.Printf("get listenPath err: %v\n", err)
		return
	}
	log.Println("listenPath: ", appConfig.ListenPath)

	appConfig.IsAutoReLoad, err = conf.GetBool("isAutoReLoad")
	if err != nil {
		log.Printf("get isAutoReLoad err: %v\n", err)
		return
	}
	log.Println("isAutoReLoad: ", appConfig.IsAutoReLoad)

	AppConfigManager.Config.Store(&appConfig)
	log.Println("first load success.")

}
