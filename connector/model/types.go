package model

type Instance struct {
	ID         string
	Name       string        // name of the instance
	Plugin     string        // name of connector
	Connection ConnInterface `json:"-"`
	Config     interface{}
}

type InstanceView struct {
	ID     string
	Name   string // name of the instance
	Plugin string // name of connector
	State  string
	Config interface{}
}

type Plugin struct {
	Constructor Constructor `json:"-"`
	Config      interface{}
}

type ConnInterface interface {
	LoadConfig(config interface{}) error
	Init() error
	Stop()
	GetConnection() interface{}
	GetState() string
}

// plugin registry

type Constructor func(name string, config interface{}) ConnInterface
