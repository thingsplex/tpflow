package websocket

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
)

type Connector struct {
	name         string
	state        string
	config       ConnectorConfig
	msgStreams   map[string]MsgPipeline
	msgTransport *websocket.Conn
}

type ConnectorConfig struct {
	WsServerURI string
	WsToken     string
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
	return &con
}

func (conn *Connector) LoadConfig(config interface{}) error {
	return mapstructure.Decode(config, &conn.config)
}

func (conn *Connector) GetConfig() interface{} {
	return conn.config
}

func (conn *Connector) SetDefaults() bool {
	return false
}

func (conn *Connector) Init() error {
	conn.state = "INIT_FAILED"
	log.Info("<WsConn> Initializing WS outbound connection")
	var err error
	conn.msgTransport, _, err = websocket.DefaultDialer.Dial(conn.config.WsServerURI, nil)
	if err != nil {
		log.Error("<WsConn> Connection error : ", err)
		return err
	}

	log.Info("<WsConn>  WS connection established")
	if err != nil {
		log.Error("<WsConn> Error connecting to broker : ", err)
	} else {
		conn.state = "RUNNING"
	}
	return err
	//conn.msgTransport.SetMessageHandler(conn.onMqttMessage)
}

func (conn *Connector) Stop() {
	conn.state = "STOPPED"
	conn.msgTransport.Close()
}

func (conn *Connector) GetConnection() interface{} {
	return conn.msgTransport
}

func (conn *Connector) GetState() string {
	return conn.state
}
