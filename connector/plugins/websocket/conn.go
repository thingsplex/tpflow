package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"net/http"
	"time"
)

type Connector struct {
	name          string
	state         string
	config        ConnectorConfig
	msgStreams    map[string]MsgStream
	msgTransport  *websocket.Conn
	maxRetryCount int
}

type ConnectorConfig struct {
	WsServerURI string
	WsToken     string
	Headers     map[string]string
}

type WsEvent struct {
	Event           string
	PayloadEncoding byte
	Payload         []byte
}

type MsgStream chan WsEvent

func NewConnectorInstance(name string, config interface{}) model.ConnInterface {
	con := Connector{name: name, maxRetryCount: 30}
	con.msgStreams = make(map[string]MsgStream)
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
	err := conn.connect()
	if err != nil {
		return err
	}
	go conn.StartMsgProcessingLoop()
	return err
}

func (conn *Connector) connect() error {
	log.Info("<WsConn> Initializing WS outbound connection")
	var err error
	headers := http.Header{}
	for k, v := range conn.config.Headers {
		headers[k] = []string{v}
	}
	conn.msgTransport, _, err = websocket.DefaultDialer.Dial(conn.config.WsServerURI, headers)
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
}

func (conn *Connector) Stop() {
	conn.state = "STOPPED"
	conn.msgTransport.Close()
}

func (conn *Connector) GetConnection() interface{} {
	return conn
}

func (conn *Connector) GetState() string {
	return conn.state
}

func (conn *Connector) StartMsgProcessingLoop() {
	// 1. Read binary message from connection
	// 2. Parse message into JSON or other serialization format
	// 3. Save message into the intermediate format and send over channels to all registered nodes (subscribers) .
	errorCounter := 0
	for {
		msgType, msg, err := conn.msgTransport.ReadMessage()
		if err != nil {
			log.Error("<WsConn> WS Read error :", err)
			// TODO : implement retry
			conn.state = "FAILED"
			errorCounter++
			time.Sleep(time.Second * 5)
			if errorCounter > conn.maxRetryCount {
				log.Error("<WsConn> Too many errors , failing the connector", err)
				conn.state = "FAILED"
				break
			}
			conn.connect()
			continue
		} else {
			log.Debug("<WsConn> WS Read message :", string(msg))
			errorCounter = 0
		}
		log.Debug("<WsConn> Number of subscribers : ", len(conn.msgStreams))
		// broadcasting message to all connected nodes
		for _, msgStream := range conn.msgStreams {
			select {
			case msgStream <- WsEvent{Event: "", Payload: msg, PayloadEncoding: byte(msgType)}:
				log.Debug("<WsConn> Msg sent to msgStream :")
			default:
				log.Warning("<WsConn> Message stream is full")
			}
		}
	}
	conn.state = "CRASHED"

}

func (conn *Connector) RegisterTrigger(nodeID string) MsgStream {
	stream := make(MsgStream, 5)
	conn.msgStreams[nodeID] = stream
	return stream
}

func (conn *Connector) UnregisterTrigger(nodeID string) {
	delete(conn.msgStreams, nodeID)
}

func (conn *Connector) Publish(wsPayloadType int, payload []byte) error {
	return conn.msgTransport.WriteMessage(wsPayloadType, payload)
}
