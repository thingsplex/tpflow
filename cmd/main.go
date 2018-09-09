package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	tpflow "github.com/alivinco/tpflow"
	fapi "github.com/alivinco/tpflow/api"
	"github.com/alivinco/tpflow/flow"
	"github.com/alivinco/tpflow/registry"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
	flowManager.GetConnectorRegistry().AddInstance("thing_registry","thing_registry","RUNNING",thingRegistryStore)
	flowManager.InitMessagingTransport()
	err = flowManager.LoadAllFlowsFromStorage()
	if err != nil {
		log.Error("Can't load Flows from storage . Error :", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	fapi.NewContextApi(flowManager.GetGlobalContext(),e)
	fapi.NewFlowApi(flowManager,e)
	fapi.NewRegistryApi(thingRegistryStore,e)


	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:4200", "http:://localhost:8082","http:://localhost:8083"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))

	log.Info("<main> Started")

	e.Logger.Debug(e.Start(":8083"))

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	_ = <-c
	thingRegistryStore.Disconnect()

	log.Info("<main> Application is terminated")
	
}
