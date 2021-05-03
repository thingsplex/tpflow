package http

import (
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/utils"
	"net/http"
	"sync"
	"time"
)

type Connector struct {
	name               string
	state              string
	config             ConnectorConfig
	server             *http.Server
	router             *mux.Router
	flowStreamMutex    sync.RWMutex
	flowStreamRegistry map[string]flowStream
	liveRequests       sync.Map
}

type ConnectorConfig struct {
	BindAddress string
}

type RequestEvent struct {
	HttpRequest *http.Request
	RequestId   int32
}

type flowStream struct {
	reqChannel chan RequestEvent
	isSync     bool
}

type liveRequest struct {
	respWriter     http.ResponseWriter
	startTime      time.Time
	responseSignal chan bool
}

func NewConnectorInstance(name string, config interface{}) model.ConnInterface {
	con := Connector{name: name}
	con.LoadConfig(config)
	con.Init()
	return &con
}

func (conn *Connector) LoadConfig(config interface{}) error {
	err := mapstructure.Decode(config, &conn.config)
	if conn.config.BindAddress == "" {
		conn.config.BindAddress = ":9090"
	}
	return err
}

func (conn *Connector) Init() error {
	var err error
	conn.state = "INIT_FAILED"
	log.Info("<HttpConn> Initializing HTTP router.")

	conn.server = &http.Server{Addr: conn.config.BindAddress}
	conn.router = mux.NewRouter()
	conn.router.HandleFunc("/flow/{id}", conn.flowRouter)
	conn.server.Handler = conn.router
	conn.flowStreamRegistry = map[string]flowStream{}
	go conn.server.ListenAndServe()
	conn.state = "RUNNING"
	log.Info("<HtppConn> HTTP router has been created ")
	return err
}

// flowRouter is invoked by HTTP server ,and it returns response to caller
func (conn *Connector) flowRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowId := vars["id"]
	conn.flowStreamMutex.RLock()
	log.Debug("<httpConn> Number of registered flows = ", len(conn.flowStreamRegistry))

	stream, ok := conn.flowStreamRegistry[flowId]
	conn.flowStreamMutex.RUnlock()
	log.Debug("<HttpConn> New HTTP request for flow ", flowId)
	if !ok {
		log.Debug("<HttpConn> no path for ", flowId)
		return
	}

	var reqId int32
	var responseSignal chan bool
	if stream.isSync {
		responseSignal = make(chan bool)
		reqId = utils.GenerateRandomNumber()
		conn.liveRequests.Store(reqId, liveRequest{respWriter: w, startTime: time.Now(),responseSignal:responseSignal  })
	}

	stream.reqChannel <- RequestEvent{RequestId: reqId, HttpRequest: r}

	if stream.isSync {
		<-responseSignal
	}

	log.Debug("<HttpConn> http transaction completed. ")
}

func (conn *Connector) RegisterFlow(flowId string, isSync bool, reqChannel chan RequestEvent) {
	if reqChannel == nil {
		return
	}
	conn.flowStreamMutex.Lock()
	conn.flowStreamRegistry[flowId] = flowStream{
		reqChannel: reqChannel,
		isSync:     isSync,
	}
	conn.flowStreamMutex.Unlock()
	log.Debug("<HttpConn> Registered flow with id = ", flowId)
}

func (conn *Connector) UnregisterFlow(flowId string) {
	conn.flowStreamMutex.Lock()
	delete(conn.flowStreamRegistry, flowId)
	conn.flowStreamMutex.Unlock()
}

func (conn *Connector) ReplyToRequest(requestId int32, body []byte,responseContentType string) {
	if requestId == 0 {
		return
	}
	wi, ok := conn.liveRequests.Load(requestId)
	if !ok {
		return
	}
	defer conn.liveRequests.Delete(requestId)
	w, ok := wi.(liveRequest)
	if !ok {
		return
	}
	log.Debug("<httpConn> Sending http reply , payload size = ", len(body))
	headers := w.respWriter.Header()
	headers.Set("Content-Type", responseContentType)
	w.respWriter.Write(body)
	w.responseSignal <- true
}

func (conn *Connector) Stop() {
	conn.server.Close()
	conn.state = "STOPPED"
}

func (conn *Connector) GetConnection() interface{} {
	return conn
}

func (conn *Connector) GetState() string {
	return conn.state
}
