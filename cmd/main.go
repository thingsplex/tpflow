package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow"
	fapi "github.com/thingsplex/tpflow/api"
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/flow"
	"github.com/thingsplex/tpflow/gstate"
	"github.com/thingsplex/tpflow/registry/integration/fimpcore"
	"github.com/thingsplex/tpflow/registry/storage"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"

	//_ "net/http/pprof"
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

// overrides configs
func processEnvVars(configs *tpflow.Configs) {
	if mqttUrl := os.Getenv("MQTT_URI"); mqttUrl != "" {
		configs.MqttServerURI = mqttUrl
	}
	if mqttUsername := os.Getenv("MQTT_USERNAME"); mqttUsername != "" {
		configs.MqttUsername = mqttUsername
	}
	if mqttPass := os.Getenv("MQTT_PASSWORD"); mqttPass != "" {
		configs.MqttPassword = mqttPass
	}
}

func InitApiMqttTransport(config tpflow.Configs) (*fimpgo.MqttTransport, error) {
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
	return msgTransport, err
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
	var registryBackend string
	processEnvVars(&configs)
	if configs.RegistryBackend == "" {
		registryBackend = "vinculum"
	} else {
		registryBackend = configs.RegistryBackend
	}

	SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting Thingsplex-Flow----------------")

	//---------FLOW------------------------
	log.Info("<main> Starting Flow manager")
	flowManager, err := flow.NewManager(configs)
	if err != nil {
		log.Error("Can't Init Flow manager . Error :", err)
	}

	var assetRegistry storage.RegistryStorage
	var fimpConConf fimpmqtt.ConnectorConfig
	//---------<THINGS REGISTRY>-------------
	if registryBackend == "vinculum" {
		// Loading registry from vinculum using configurations from fimpmqtt connector
		conInst, ok := flowManager.GetConnectorRegistry().GetInstance("fimpmqtt").Connection.(*fimpmqtt.Connector)
		if !ok {
			log.Error("Can't load connector")
		}
		fimpConConf, ok = conInst.GetConfig().(fimpmqtt.ConnectorConfig)
		if ok {
			assetRegistry = storage.NewVinculumRegistryStore(&fimpConConf, configs.IsDevMode)
			assetRegistry.Connect()
		} else {
			log.Error("Can't load fimpqtt configs")
		}

	} else if registryBackend == "local" {
		log.Info("<main>-------------- Starting service assetRegistry ")
		assetRegistry = storage.NewThingRegistryStore(configs.RegistryDbFile)
		log.Info("<main> Started ")
	}
	log.Info("<main>-------------- Starting service assetRegistry integration ")
	regMqttIntegr := fimpcore.NewMqttIntegration(&configs, assetRegistry)
	regMqttIntegr.InitMessagingTransport()
	log.Info("<main> Started ")
	//---------</THINGS REGISTRY>------------

	flowManager.GetConnectorRegistry().AddConnection("thing_registry", "thing_registry", "thing_registry", assetRegistry, nil)
	err = flowManager.LoadAllFlowsFromStorage()
	if err != nil {
		log.Error("Can't load Flows from storage . Error :", err)
	}

	ctxApi := fapi.NewContextApi(flowManager.GetGlobalContext())
	flowApi := fapi.NewFlowApi(flowManager, &configs)
	regApi := fapi.NewRegistryApi(assetRegistry)

	stateTracker := gstate.NewGlobalStateTracker(flowManager.GetGlobalContext(), assetRegistry)
	stateTracker.InitFimpStateExtractor(&fimpConConf)
	connectorReg := flowManager.GetConnectorRegistry()

	var httpSrvConnector *http.Connector
	if connectorReg != nil {
		inst := connectorReg.GetInstance("httpserver")
		httpSrvConnector = inst.Connection.(*http.Connector)
		httpSrvConnector.SetAssetRegistry(assetRegistry)
		httpSrvConnector.SetFlowContext(flowManager.GetGlobalContext())
	}

	apiMqttTransport, err := InitApiMqttTransport(configs)

	if err == nil {
		flowApi.RegisterMqttApi(apiMqttTransport, httpSrvConnector)
		regApi.RegisterMqttApi(apiMqttTransport)
		ctxApi.RegisterMqttApi(apiMqttTransport)
	}

	log.Info("<main> Started")
	//<< PPPROF >>
	//if configs.EnableProfiler {
	//	go func() {
	//		// http://cube.local:6060/debug/pprof/
	//		//  go tool pprof  http://cube.local:6060/debug/pprof/heap
	//		http.ListenAndServe(":6060", nil)
	//	}()
	//}

	//e.Logger.Debug(e.Start(":8083"))

	//c := make(chan os.Signal, 1)
	//signal.Notify(c)
	//_ = <-c
	runtime.GC()
	select {}

	assetRegistry.Disconnect()

	log.Info("<main> Application is terminated")

}
