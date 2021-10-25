package timeseries

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/connector/plugins/fimpmqtt"
)

type Connector struct {
	name         string
	state        string
	config       fimpmqtt.ConnectorConfig
	msgTransport *fimpgo.MqttTransport
	syncClient   *fimpgo.SyncClient
	requestTopic string
	responseTopic string

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
	log.Info("<TsFimpConn> Initializing fimp MQTT client.")
	clientId := conn.config.MqttClientIdPrefix + "_timeseries"
	conn.msgTransport = fimpgo.NewMqttTransport(conn.config.MqttServerURI, clientId, conn.config.MqttUsername, conn.config.MqttPassword, true, 0, 0)
	conn.msgTransport.SetGlobalTopicPrefix(conn.config.MqttTopicGlobalPrefix)
	err := conn.msgTransport.Start()
	log.Info("<TsFimpConn> Mqtt transport connected")
	if err != nil {
		log.Error("<TsFimpConn> Error connecting to broker : ", err)
	} else {
		conn.state = "RUNNING"
	}
	conn.syncClient = fimpgo.NewSyncClient(conn.msgTransport)
	conn.requestTopic = "pt:j1/mt:cmd/rt:app/rn:ecollector/ad:1"
	conn.responseTopic = "pt:j1/mt:rsp/rt:app/rn:tpflow/ad:1"
	conn.msgTransport.Subscribe(conn.responseTopic)
	return err
}

func (conn *Connector) Stop() {
	conn.state = "STOPPED"
	conn.msgTransport.Stop()
}

func (conn *Connector) GetConnection() interface{} {
	return conn
}

func (conn *Connector) GetState() string {
	return conn.state
}

// Ecollector API

func (conn *Connector) QueryDataPoints(query string) (*Result,error) {
	reqMsg := fimpgo.NewStrMapMessage("cmd.tsdb.query","ecollector", map[string]string{"query":query},nil,nil,nil)
	return conn.sendRequest(reqMsg)
}

func (conn *Connector) GetDataPoints(request *GetDataPointsRequest) (*Result,error) {
	reqMsg := fimpgo.NewObjectMessage("cmd.tsdb.get_data_points","ecollector",request ,nil,nil,nil)
	return conn.sendRequest(reqMsg)
}

func (conn *Connector) GetEnergyDataPoints(request *GetDataPointsRequest) (*Result,error) {
	reqMsg := fimpgo.NewObjectMessage("cmd.tsdb.get_energy_data_points","ecollector",request ,nil,nil,nil)
	return conn.sendRequest(reqMsg)
}

func (conn *Connector) WriteDataPoints(request *WriteDataPointsRequest) error {
	reqMsg := fimpgo.NewObjectMessage("cmd.tsdb.write_data_points","ecollector",request ,nil,nil,nil)
	return conn.msgTransport.PublishToTopic(conn.requestTopic,reqMsg)
}

func (conn *Connector) sendRequest(reqMsg *fimpgo.FimpMessage)(*Result,error) {
	reqMsg.ResponseToTopic = conn.responseTopic
	respMsg , err := conn.syncClient.SendReqRespFimp(conn.requestTopic,conn.responseTopic,reqMsg , 20,false)
	if err != nil {
		return nil,err
	}
	resp := Response{}
	err = respMsg.GetObjectValue(&resp)
	if err != nil {
		return nil,err
	}
	if len(resp.Results)>0 {
		return &resp.Results[0],nil
	} else {
		log.Debugf("<TsFimpConn> Empty response")
		return nil, err
	}
}
