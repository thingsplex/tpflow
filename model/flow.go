package model

import (
	"github.com/futurehomeno/fimpgo"
	"time"
)

type MsgPipeline chan Message

type FlowRunner func(ReactorEvent)

type Message struct {
	AddressStr string
	Address    fimpgo.Address
	Payload    fimpgo.FimpMessage
	RawPayload []byte
	Header     map[string]string
	CancelOp   bool // if true , listening end should close all operations
}

type ReactorEvent struct {
	Msg              Message
	Err              error
	TransitionNodeId NodeID
	SrcNodeId        NodeID
}

type FlowMeta struct {
	Id                string // Instance id . Is different for every instance
	ClassId           string // Class id , all instances share the same ClassId
	Author            string
	Version           int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Name              string
	Group             string
	Description       string
	Nodes             []MetaNode
	Settings          map[string]string
	IsDisabled        bool
	IsDefault         bool // default flows are read only and can't be deleted
	ParallelExecution string // keep_first , keep_last , parallel
}

const (
	SIGNAL_STOP                = 1
	SIGNAL_TERMINATE_WAITING   = 2 // Signal to terminate all waiting nodes but not trigger nodes
	ParallelExecutionKeepFirst = "keep_first"
	ParallelExecutionKeepLast  = "keep_last"
	ParallelExecutionParallel  = "parallel"
)

type FlowOperationalContext struct {
	FlowMeta                    *FlowMeta
	FlowId                      string
	IsFlowRunning               bool
	State                       string
	TriggerControlSignalChannel chan int // the channel should be used to stop all waiting nodes .
	NodeControlSignalChannel    chan int
	NodeIsReady                 chan bool // Flow should notify message router when next node is ready to process new message .
	StoragePath                 string
	ExtLibsDir                  string
}

type FlowStatsReport struct {
	CurrentNodeId          NodeID
	CurrentNodeLabel       string
	NumberOfNodes          int
	NumberOfTriggers       int
	NumberOfActiveTriggers int
	NumberOfActiveSubflows int
	State                  string
	StartedAt              time.Time
	WaitingSince           time.Time
	LastExecutionTime      int64
}
