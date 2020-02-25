package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	fapi "github.com/thingsplex/tpflow/api"
	"github.com/thingsplex/tpflow/flow"
	"github.com/thingsplex/tpflow/registry/integration/fimpcore"
	"github.com/thingsplex/tpflow/registry/storage"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

// SetupLog configures default logger
// Supported levels : info , degug , warn , error
func SetupLog(logfile string, level string, logFormat string) {
	if logFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05.999"})
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, ForceColors: true, TimestampFormat: "2006-01-02T15:04:05.999"})
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

func InitApiMqttTransport(config tpflow.Configs) (*fimpgo.MqttTransport,error) {
	log.Info("<main> Initializing fimp MQTT client.")
	clientId := config.MqttClientIdPrefix + "tpflow_api"
	msgTransport := fimpgo.NewMqttTransport(config.MqttServerURI, clientId, config.MqttUsername, config.MqttPassword, true, 1, 1)
	msgTransport.SetGlobalTopicPrefix(config.MqttTopicGlobalPrefix)
	err := msgTransport.Start()
	log.Info("<main> Mqtt transport connected")
	if err != nil {
		log.Error("<main> Error connecting to broker : ", err)
	} else {

	}
	return msgTransport,err
}

func main() {
	configs := tpflow.Configs{}
	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file")
	flag.Parse()
	if configFile == "" {
		configFile = "./config.json"
	} else {
		fmt.Println("Loading configs from file ", configFile)
	}
	configFileBody, err := ioutil.ReadFile(configFile)
	err = json.Unmarshal(configFileBody, &configs)
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}
	registryBackend := "vinculum"

	SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting Thingsplex-Flow----------------")
	var registry storage.RegistryStorage
	//---------THINGS REGISTRY-------------
	if registryBackend == "vinculum" {
		registry = storage.NewVinculumRegistryStore(&configs)
		registry.Connect()
	}else if registryBackend == "local" {
		log.Info("<main>-------------- Starting service registry ")
		registry = storage.NewThingRegistryStore(configs.RegistryDbFile)
		log.Info("<main> Started ")
	}
	log.Info("<main>-------------- Starting service registry integration ")
	regMqttIntegr := fimpcore.NewMqttIntegration(&configs, registry)
	regMqttIntegr.InitMessagingTransport()
	log.Info("<main> Started ")

	//---------FLOW------------------------
	log.Info("<main> Starting Flow manager")
	flowManager, err := flow.NewManager(configs)
	if err != nil {
		log.Error("Can't Init Flow manager . Error :", err)
	}
	flowManager.GetConnectorRegistry().AddConnection("thing_registry", "thing_registry", "thing_registry", registry)
	err = flowManager.LoadAllFlowsFromStorage()
	if err != nil {
		log.Error("Can't load Flows from storage . Error :", err)
	}

	ctxApi := fapi.NewContextApi(flowManager.GetGlobalContext(), nil)
	flowApi := fapi.NewFlowApi(flowManager, nil,&configs)
	regApi := fapi.NewRegistryApi(registry, nil)

	apiMqttTransport,err := InitApiMqttTransport(configs)

	if err == nil {
		flowApi.RegisterMqttApi(apiMqttTransport)
		regApi.RegisterMqttApi(apiMqttTransport)
		ctxApi.RegisterMqttApi(apiMqttTransport)
	}

	log.Info("<main> Started")
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
	//e.Logger.Debug(e.Start(":8083"))

	//c := make(chan os.Signal, 1)
	//signal.Notify(c)
	//_ = <-c
    runtime.GC()
	select {}

	registry.Disconnect()

	log.Info("<main> Application is terminated")

}
