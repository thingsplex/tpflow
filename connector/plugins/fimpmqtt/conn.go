package fimpmqtt

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/alivinco/tpflow/connector/model"
	"github.com/labstack/gommon/log"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	name         string
	state        string
	config       ConnectorConfig
	msgStreams   map[string]MsgPipeline
	msgTransport *fimpgo.MqttTransport
}

type ConnectorConfig struct {
	MqttServerURI         string
	MqttUsername          string
	MqttPassword          string
	MqttTopicGlobalPrefix string
	MqttClientIdPrefix    string
}

type MsgPipeline chan Message

type Message struct {
	AddressStr string
	Address    fimpgo.Address
	Payload    fimpgo.FimpMessage
	RawPayload []byte
	Header     map[string]string
	CancelOp   bool // if true , listening end should close all operations
}

func NewConnectorInstance(name string, config interface{}) model.ConnInterface {
	con := Connector{name: name}
	con.LoadConfig(config)
	con.Init()
	return &con
}

func (conn *Connector) LoadConfig(config interface{}) error {
	return mapstructure.Decode(config, &conn.config)
}

func (conn *Connector) Init() error {
	conn.state = "INIT_FAILED"
	log.Info("<FimpConn> Initializing fimp MQTT client.")
	clientId := conn.config.MqttClientIdPrefix + "flow_manager"
	conn.msgTransport = fimpgo.NewMqttTransport(conn.config.MqttServerURI, clientId, conn.config.MqttUsername, conn.config.MqttPassword, true, 1, 1)
	conn.msgTransport.SetGlobalTopicPrefix(conn.config.MqttTopicGlobalPrefix)
	err := conn.msgTransport.Start()
	log.Info("<FimpConn> Mqtt transport connected")
	if err != nil {
		log.Error("<FimpConn> Error connecting to broker : ", err)
	} else {
		conn.state = "RUNNING"
	}
	return err
	//conn.msgTransport.SetMessageHandler(conn.onMqttMessage)
}

func (conn *Connector) Stop() {
	conn.state = "STOPPED"
	conn.msgTransport.Stop()

}

// Returns
func (conn *Connector) GetConnection() interface{} {
	return conn.msgTransport
}

func (conn *Connector) GetState() string {
	return conn.state
}
