package http

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/utils"
	"html/template"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

var (
	brUpgrader = websocket.Upgrader{
		Subprotocols: []string{},
		HandshakeTimeout: time.Second*20,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Connector struct {
	name               string
	state              string
	config             ConnectorConfig
	server             *http.Server
	router             *mux.Router
	flowStreamMutex    sync.RWMutex
	flowStreamRegistry map[string]flowStream
	liveConnections    sync.Map // map of type liveConnection
	isServerStarted    bool // for lazy loading
}

type ConnectorConfig struct {
	BindAddress string
}

type RequestEvent struct {
	HttpRequest *http.Request
	RequestId   int32
	IsWsMsg     bool
	Payload     []byte
}

type flowStream struct {
	reqChannel    chan RequestEvent // channel is used to send message from HTTP/WS to flow trigger node
	isSync        bool
	IsWs          bool
	isPublishOnly bool // true - mean that flow will only send messages to client connection but is not interested in receiving events . false - flow  has trigger
	FlowIdAlias   string // alias to flowId , can be used in url instead of flowId
	Name          string // human readable name
}

type liveConnection struct {
	flowId         string
	respWriter     http.ResponseWriter // channel is used by HTTP/WS action node for sending response to http request
	startTime      time.Time
	responseSignal chan bool
	isWs           bool
	wsConn         *websocket.Conn
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
	log.Info("<HttpConn> Configuring HTTP router.")
	conn.server = &http.Server{Addr: conn.config.BindAddress}
	conn.router = mux.NewRouter()
	conn.router.HandleFunc("/index", conn.index)
	conn.router.HandleFunc("/flow/{id}/rest", conn.httpFlowRouter)
	conn.router.HandleFunc("/flow/{id}/ws", conn.wsFlowRouter)
	conn.server.Handler = conn.router
	conn.flowStreamRegistry = map[string]flowStream{}
	conn.state = "RUNNING"
	log.Info("<HtppConn> HTTP router configured ")
	return err
}

func (conn *Connector) StartHttpServer() {
	log.Info("<HttpConn> Starting HTTP server.")
	go conn.server.ListenAndServe()
	conn.isServerStarted = true
}

// httpFlowRouter is invoked by HTTP server ,and it returns response to caller
func (conn *Connector) httpFlowRouter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowId := vars["id"]
	conn.flowStreamMutex.RLock()
	//log.Debug("<httpConn> Number of registered flows = ", len(conn.flowStreamRegistry))
	stream, ok := conn.flowStreamRegistry[flowId]
	conn.flowStreamMutex.RUnlock()
	log.Debug("<HttpConn> New HTTP request for flow ", flowId)
	if !ok {
		log.Debug("<HttpConn> no path for ", flowId)
		return
	}

	var reqId int32
	var responseSignal chan bool
	responseSignal = make(chan bool)
	reqId = utils.GenerateRandomNumber()
	conn.liveConnections.Store(reqId, liveConnection{respWriter: w, startTime: time.Now(),responseSignal:responseSignal  })

	if stream.reqChannel != nil {
		stream.reqChannel <- RequestEvent{RequestId: reqId, HttpRequest: r}
	}

	if responseSignal !=nil {
		<-responseSignal
	}

	log.Debug("<HttpConn> http transaction completed. ")
}

// wsFlowRouter is invoked by HTTP server to convert HTTP request to WS request.Method is blocked until connection is alive
func (conn *Connector) wsFlowRouter(w http.ResponseWriter, r *http.Request) {
	reqId := utils.GenerateRandomNumber()
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> WS connection failed with PANIC")
			log.Error(string(debug.Stack()))
		}
		conn.liveConnections.Delete(reqId)
		log.Debug("<HttpConn> WS connection closed")
	}()
	vars := mux.Vars(r)
	flowId := vars["id"]

	conn.flowStreamMutex.RLock()
	stream, ok := conn.flowStreamRegistry[flowId]
	conn.flowStreamMutex.RUnlock()
	log.Debug("<HttpConn> New WS request for flow ", flowId)
	if !ok {
		log.Debug("<HttpConn> no path for ", flowId)
		return
	}
	if !stream.IsWs {
		log.Info("<HttpConn> The stream doesn't support WS capabilities")
		return
	}

	ws, err := brUpgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("<HttpConn> Can't upgrade to WS . Error:", err)
		return
	}

	conn.liveConnections.Store(reqId, liveConnection{respWriter: w, startTime: time.Now(),isWs: true,wsConn: ws,flowId: flowId})

	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Error("<HttpConn> WS Read error :", err)
			break
		} else if msgType == websocket.TextMessage {
			if stream.isPublishOnly {
				continue
			}else {
				if stream.reqChannel != nil {
					stream.reqChannel <- RequestEvent{RequestId: 0, HttpRequest: r,IsWsMsg: true, Payload: msg}
				}
			}
		} else if msgType == websocket.BinaryMessage {
			log.Debug("<HttpConn> New binary ws message")
		} else {
			log.Debug("<HttpConn> Message of type = ", msgType)
		}
	}

}

func (conn *Connector) index(w http.ResponseWriter, r *http.Request) {
	const indexTemplate = `
    <html lang="en">
    <head> <title> TPflow http/ws endpoints </title> </head>
    <body>
    <p>List of endpoints</p>
    {{ range $index, $element := . }} 
      {{ if $element.IsWs }}
        <p> Websocket  : <a href="/flow/{{ $index }}/ws"> /flow/{{ $index }}/ws </a> - {{ $element.Name }} </p> 
      {{ else }} 
		<p> Http : <a href="/flow/{{ $index }}/rest"> /flow/{{ $index }}/rest </a> - {{ $element.Name }} </p> 
      {{ end }} 
    {{ end }}
    </body>
    </html>
    `
	t, err := template.New("foo").Parse(indexTemplate)
	if err != nil {
		return
	}
	var out bytes.Buffer
	err = t.Execute(&out, conn.flowStreamRegistry)
	w.Write(out.Bytes())
}

// RegisterFlow registers node of the flow
func (conn *Connector) RegisterFlow(flowId string, isSync bool,isWs bool,publishOnly bool, reqChannel chan RequestEvent,alias,name string ) {
	if reqChannel == nil {
		return
	}
	// Lazy loading. Server started only after user creates first flow that using HTTP trigger or action nodes.
	if !conn.isServerStarted {
		conn.StartHttpServer()
	}
	conn.flowStreamMutex.Lock()
	conn.flowStreamRegistry[flowId] = flowStream{
		reqChannel:    reqChannel,
		isSync:        isSync,
		IsWs:          isWs,
		isPublishOnly: publishOnly,
		FlowIdAlias:  alias,
		Name: name,
	}
	conn.flowStreamMutex.Unlock()
	log.Debug("<HttpConn> Registered flow with id = ", flowId)
}

func (conn *Connector) UnregisterFlow(flowId string) {
	conn.liveConnections.Range(func(key, value interface{}) bool {
		// republishing messages to all connected clients
		lConn,ok := value.(liveConnection)
		if !ok {
			return true
		}
		if lConn.flowId == flowId && lConn.isWs {
			lConn.wsConn.Close()
			conn.liveConnections.Delete(key)
			log.Debug("<HttpConn> Connection deleted , id=",key)
		}

		return true
	})

	conn.flowStreamMutex.Lock()
	delete(conn.flowStreamRegistry, flowId)
	conn.flowStreamMutex.Unlock()

	if len(conn.flowStreamRegistry)==0 {
		conn.server.Close()
		conn.isServerStarted = false
		log.Debug("<HttpConn> No HTTP flows left , shutting down HTTP server")
	}
}

func (conn *Connector) ReplyToRequest(requestId int32, payload []byte,responseContentType string) {
	if requestId == 0 {
		return
	}
	wi, ok := conn.liveConnections.Load(requestId)
	if !ok {
		return
	}
	defer conn.liveConnections.Delete(requestId)
	lConn, ok := wi.(liveConnection)
	if !ok {
		return
	}
	if lConn.isWs {
		// Connection was initiated using WS trigger
		err := lConn.wsConn.WriteMessage(websocket.TextMessage,payload)
		if err != nil {
			log.Debug("<httpConn> Message forwarded to client")
		}
	}else{
		if payload != nil {
			log.Debug("<httpConn> Sending http reply , Payload size = ", len(payload))
			headers := lConn.respWriter.Header()
			headers.Set("Content-Type", responseContentType)
			lConn.respWriter.Write(payload)
		}
		lConn.responseSignal <- true
	}

}

// PublishWs must be used to publish messages to live WS connection , given that flow is not triggered by the same connection and there is not WS trigger.
func (conn *Connector) PublishWs(flowId string,payload []byte) {
	conn.liveConnections.Range(func(key, value interface{}) bool {
		// republishing messages to all connected clients
		lConn,ok := value.(liveConnection)
		if !ok {
			return true
		}
		if lConn.flowId == flowId {
			// TODO : Research is the operation must be executed in async way to avoid blocking , which can happen if client consumes with different speed.
			lConn.wsConn.SetWriteDeadline(time.Now().Add(time.Second*10))
			err := lConn.wsConn.WriteMessage(websocket.TextMessage,payload)
			if err == nil {
				log.Debug("<httpConn> Message forwarded to client")
			}else {
				log.Info("<httpClient> Can't write to WS connection. Err:",err.Error())
			}
		}
		return true
	})
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
