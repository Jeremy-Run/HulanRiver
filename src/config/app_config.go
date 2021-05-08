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
