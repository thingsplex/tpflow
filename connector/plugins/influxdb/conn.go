package influxdb

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/tpflow/connector/model"
	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/mitchellh/mapstructure"
)

type Connector struct {
	name    string
	influxC influx.Client
	state   string
	config  ConnectorConfig
}

type ConnectorConfig struct {
	Address             string
	Username            string
	Password            string
	Db                  string
	RetentionPolicyName string
	RetentionDuration   string
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
	var err error
	conn.state = "INIT_FAILED"
	log.Info("<InfluxdbConn> Initializing influx client.")
	conn.influxC, err = influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     conn.config.Address, //"http://localhost:8086",
		Username: conn.config.Username,
		Password: conn.config.Password,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
		return err
	}
	// Creating database
	log.Info("<InfluxdbConn> Setting up database")
	q := influx.NewQuery(fmt.Sprintf("CREATE DATABASE %s", conn.config.Db), "", "")
	if response, err := conn.influxC.Query(q); err == nil && response.Error() == nil {
		log.Infof("<InfluxdbConn> Database %s was created with status :%s", conn.config.Db, response.Results)
	} else {
		log.Error("<InfluxdbConn> Database init failed . Error:", err)
		return err
	}
	// Setting up retention policies
	log.Info("<InfluxdbConn>  Setting up retention policies")
	q = influx.NewQuery(fmt.Sprintf("CREATE RETENTION POLICY %s ON %s DURATION %s REPLICATION 1", conn.config.RetentionPolicyName, conn.config.Db, conn.config.RetentionDuration), conn.config.Db, "")
	if response, err := conn.influxC.Query(q); err == nil && response.Error() == nil {
		log.Infof("<InfluxdbConn> Retention policy %s was created with status :%s", conn.config.RetentionPolicyName, response.Results)
	} else {
		log.Errorf("<InfluxdbConn> Configuration of retention policy %s failed with status : %s ", conn.config.RetentionPolicyName, response.Error())
	}
	conn.state = "RUNNING"
	return err

}

func (conn *Connector) Stop() {
	conn.state = "STOPPED"
	conn.influxC.Close()

}

func (conn *Connector) GetConnection() interface{} {
	return conn.influxC
}

func (conn *Connector) GetState() string {
	return conn.state
}
