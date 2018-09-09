package connector


type Instance struct {
	Name string  // name of the instance
	ConnectorType string  // name of connector
	State string //
	Connection interface{}

}

type ConnInterface interface {
	LoadConfig(config interface{})error
	Init()error
	Stop()
	GetConnection() interface{}
}