package model

import (
	"github.com/thingsplex/tpflow/connector"
)

type NodeID string

type MetaNode struct {
	Id                NodeID
	Type              string
	Label             string
	SuccessTransition NodeID
	TimeoutTransition NodeID
	ErrorTransition   NodeID
	Address           string
	Service           string
	ServiceInterface  string
	Config            interface{}
	Ui                interface{}
}

type NodeVariableDef struct {
	Name        string
	InMemory    bool
	IsGlobal    bool
	Type        string
}

type Node interface {
	// OnInput is invoked by flow runner. msg is original message from trigger , node can perform even mutation.
	OnInput(msg *Message) ([]NodeID, error)
	WaitForEvent(responseChannel chan ReactorEvent)
	GetMetaNode() *MetaNode
	GetNextSuccessNodes() []NodeID
	GetNextErrorNode() NodeID
	GetNextTimeoutNode() NodeID
	LoadNodeConfig() error
	IsStartNode() bool
	IsMsgReactorNode() bool
	IsReactorRunning() bool
	// Init is invoked when node is started
	Init() error
	// Cleanup is invoked when node is stopped
	Cleanup() error
	SetConnectorRegistry(connectorRegistry *connector.Registry)
	SetFlowRunner(runner FlowRunner)
	//GetAdapterInstance()

	//GetConfigs() interface{}
	//SetConfigs(configs interface{})
}
