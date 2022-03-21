package main

import (
	log "github.com/sirupsen/logrus"
	"dataframe-service/src/app/config"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/rifflock/lfshook"
	"io/ioutil"
	"time"
	"sync"
	"dataframe-service/src/app/rest"
	"dataframe-service/src/app/utils"
)

func main() {
	fmt.Println("Dataframe service")
}
func init() {
	err := config.ReadYamlConfigFile()
	if err != nil {
		log.Errorln(err)
		return
	}
	logFile := "output/log/logger.log"
	LogWriter := lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1024, // megabytes
		MaxAge:     config.GetConfig().DataframeConfig.Log.MaxAge,
		MaxBackups: config.GetConfig().DataframeConfig.Log.MaxBackups,
		Compress:   true,
	}
	writerMap := lfshook.WriterMap{
		log.DebugLevel: &LogWriter,
		log.InfoLevel:  &LogWriter,
		log.WarnLevel:  &LogWriter,
		log.ErrorLevel: &LogWriter,
	}
	log.AddHook(lfshook.NewHook(
		writerMap,
		&log.JSONFormatter{},
	))
	log.SetOutput(ioutil.Discard)
	level := config.GetConfig().DataframeConfig.Log.Level
	if len(level) > 0 {
		switch level {
		case "info":
			log.SetLevel(log.InfoLevel)
		case "warning":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		case "debug":
			log.SetLevel(log.DebugLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}

	startRestServer()
}
func startRestServer() error {
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	//Wait Groups to start both the Rest Servers
	apiServer := rest.NewAPIServer(&wg)

	if config.GetConfig().DataframeConfig.Destination.ActiveDb == "kafka"{
		brokers := config.GetConfig().DataframeConfig.Kafka.Brokers
		go utils.KafkaConsumer(brokers)
	}
	go apiServer.RunAPIServer()
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Println("Took ", elapsed)
	return nil
}
