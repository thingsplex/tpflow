package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/tpflow"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/registry"
	"github.com/alivinco/tpflow/shared"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"os/signal"
)

// SetupLog configures default logger
// Supported levels : info , degug , warn , error
func SetupLog(logfile string, level string,logFormat string ) {
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{TimestampFormat:"2006-01-02 15:04:05.999"})
	}else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true,TimestampFormat:"2006-01-02T15:04:05.999"})
	}

	logLevel, err := log.ParseLevel(level)
	if err == nil {
		log.SetLevel(logLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}

	if logfile != "" {
		l := lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    5, // megabytes
			MaxBackups: 2,
		}
		log.SetOutput(&l)
	}

}

func main() {
	configs := &model.Configs{}
	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file")
	flag.Parse()
	if configFile == "" {
		configFile = "./config.json"
	} else {
		fmt.Println("Loading configs from file ", configFile)
	}
	configFileBody, err := ioutil.ReadFile(configFile)
	err = json.Unmarshal(configFileBody, configs)
	if err != nil {
		panic("Can't load config file.")
	}

	SetupLog(configs.LogFile, configs.LogLevel,configs.LogFormat)
	log.Info("--------------Starting Thingsplex-Flow----------------")

	//---------THINGS REGISTRY-------------
	log.Info("<main>-------------- Starting service registry ")
	thingRegistryStore := registry.NewThingRegistryStore(configs.RegistryDbFile)
	log.Info("<main> Started ")
	//-------------------------------------
	//---------FLOW------------------------
	log.Info("<main> Starting Flow manager")
	flowManager, err := flow.NewManager(configs)
	if err != nil {
		log.Error("Can't Init Flow manager . Error :", err)
	}
	flowSharedResources := shared.GlobalSharedResources{Registry:thingRegistryStore}
	flowManager.SetSharedResources(flowSharedResources)
	flowManager.InitMessagingTransport()
	err = flowManager.LoadAllFlowsFromStorage()
	if err != nil {
		log.Error("Can't load Flows from storage . Error :", err)
	}
	log.Info("<main> Started")

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	_ = <-c
	thingRegistryStore.Disconnect()

	log.Info("<main> Application is terminated")
	
}
