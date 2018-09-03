package model

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

type Node interface {
	OnInput( msg *Message) ([]NodeID,error)
	// reactor nodes should publish events into the channel
	WaitForEvent(responseChannel chan ReactorEvent)
	GetMetaNode()*MetaNode
	GetNextSuccessNodes()[]NodeID
	GetNextErrorNode()NodeID
	GetNextTimeoutNode()NodeID
	LoadNodeConfig() error
	IsStartNode() bool
	IsMsgReactorNode() bool
	IsReactorRunning() bool
    ConfigureInStream(activeSubscriptions *[]string,msgInStream MsgPipeline)
    // Invoked when node is started
	Init() error
	// Invoked when node is stopped
	Cleanup() error
}







