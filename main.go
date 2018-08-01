package main

import (
	logger "github.com/sirupsen/logrus"
	"os"
	"github.com/spf13/viper"
	"os/signal"
	"github.com/tezexInfo/TezProxy"
)

const ConfigPath = "config"

func main (){

	vip := viper.New()
	vip.AddConfigPath(".")
	vip.AddConfigPath("/etc/tezproxy")
	vip.SetConfigName(ConfigPath)
	vip.ReadInConfig()


	log := logger.New()
	logger.SetFormatter(&logger.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	log.SetLevel(logger.InfoLevel)


	config := proxy.ProxyConfig{
		TezosHost:      vip.GetString("tezos.host"),
		TezosPort:      vip.GetInt("tezos.port"),
		ServerPort:     vip.GetInt("server.port"),
		Methods:        vip.GetStringSlice("proxy.whitelistedMethods"),
		ReadTimeout:    vip.GetInt("server.readTimeout"),
		WriteTimeout:   vip.GetInt("proxy.writeTimeout"),
		IdleTimeout:    vip.GetInt("proxy.idleTimeout"),
		RateLimitCount: vip.GetInt64("proxy.rateLimitCount"),
		RateLimitPeriod:    vip.GetInt("proxy.rateLimitPeriod"),
		Blocked: vip.GetStringSlice("proxy.blockedMethods"),
		DontCache: vip.GetStringSlice("proxy.dontCache"),
		CacheMaxItems: vip.GetInt("proxy.cacheMaxItems"),
	}

	proxy := proxy.NewProxy(config,log)

	proxy.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c



}

