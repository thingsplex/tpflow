package model


type Instance struct {
	ID string
	Name string  // name of the instance
	ConnectorType string  // name of connector
	State string //
	Connection interface{}

}

type Plugin struct {
	Constructor Constructor
	Config interface{}
}


type ConnInterface interface {
	LoadConfig(config interface{})error
	Init()error
	Stop()
	GetConnection() interface{}
}
// plugin registry

type Constructor func(name string,config interface{}) ConnInterface