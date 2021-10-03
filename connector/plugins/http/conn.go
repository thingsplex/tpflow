package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector/model"
	"github.com/thingsplex/tpflow/flow/context"
	model2 "github.com/thingsplex/tpflow/registry/model"
	"github.com/thingsplex/tpflow/registry/storage"
	"github.com/thingsplex/tpflow/utils"
	"github.com/thingsplex/tprelay/pkg/edge"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

var (
	brUpgrader = websocket.Upgrader{
		Subprotocols:     []string{},
		HandshakeTimeout: time.Second * 20,
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
	liveConnections    sync.Map                // map of type liveConnection
	isServerStarted    bool                    // for lazy loading
	assetRegistry      storage.RegistryStorage // this is only for exposing registry api
	flowContext        *context.Context
	tunClient          *edge.TunClient
	isTunActive        bool
}

type ConnectorConfig struct {
	BindAddress        string
	GlobalAuth         AuthConfig // None , Bearer , Basic
	IsTunEnabled       bool
	IsLocalServEnabled bool
	TunCloudEndpoint   string
	TunAddress         string // tunnel address , it can be guid of the flow engine or can be gateway id.
	TunEdgeToken       string
}

type RequestEvent struct {
	HttpRequest *http.Request
	RequestId   int64
	IsWsMsg     bool
	IsFromCloud bool
	Payload     []byte
}

type flowStream struct {
	reqChannel    chan RequestEvent // channel is used to send message from HTTP/WS to flow trigger node
	isSync        bool
	IsWs          bool
	isPublishOnly bool   // true - mean that flow will only send messages to client connection but is not interested in receiving events . false - flow  has trigger
	FlowIdAlias   string // alias to flowId , can be used in url instead of flowId
	Name          string // human readable name
	authConfig    AuthConfig
}

type liveConnection struct {
	flowId         string
	respWriter     http.ResponseWriter // channel is used by HTTP/WS action node for sending response to http request
	startTime      time.Time
	responseSignal chan bool
	isWs           bool
	isFromCloud    bool
	wsConn         *websocket.Conn
}

func NewConnectorInstance(name string, config interface{}) model.ConnInterface {
	con := Connector{name: name}
	con.LoadConfig(config)
	return &con
}

func (conn *Connector) LoadConfig(config interface{}) error {
	err := mapstructure.Decode(config, &conn.config)
	if err != nil {
		return err
	}
	if conn.state == "RUNNING" {
		if conn.isTunActive && !conn.config.IsTunEnabled {
			//Stop tunnel
			return conn.StopWsCloudTunnel()

		} else if !conn.isTunActive && conn.config.IsTunEnabled {
			//Start tunnel
			return conn.StartWsCloudTunnel()
		}
	}
	return nil
}

func (conn *Connector) GetConfig() interface{} {
	return conn.config
}

func (conn *Connector) SetDefaults() bool {
	var isChanged bool
	if conn.config.BindAddress == "" {
		conn.config.BindAddress = ":8082"
		isChanged = true
	}

	if conn.config.TunAddress == "" {
		conn.config.TunAddress = utils.GenerateUuid()
		isChanged = true
	}
	if conn.config.TunEdgeToken == "" {
		conn.config.TunEdgeToken = utils.GenerateId(16)
		isChanged = true
	}
	return isChanged
}

func (conn *Connector) Config() ConnectorConfig {
	return conn.config
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
	conn.configureInternalApi()
	conn.flowStreamRegistry = map[string]flowStream{}
	conn.state = "RUNNING"
	log.Info("<HttpConn> HTTP router configured . Is cloud tunnel enabled = ", conn.config.IsTunEnabled)
	return err
}

func (conn *Connector) SetAssetRegistry(assetRegistry storage.RegistryStorage) {
	conn.assetRegistry = assetRegistry
}

func (conn *Connector) SetFlowContext(flowContext *context.Context) {
	conn.flowContext = flowContext
}

func (conn *Connector) configureInternalApi() {
	conn.router.HandleFunc("/api/registry/devices", func(w http.ResponseWriter, r *http.Request) {
		if conn.assetRegistry != nil {
			devs, _ := conn.assetRegistry.GetExtendedDevices()
			bresp, err := json.Marshal(devs)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(bresp)
		}
	})
	conn.router.HandleFunc("/api/registry/locations", func(w http.ResponseWriter, r *http.Request) {
		if conn.assetRegistry != nil {
			locs, _ := conn.assetRegistry.GetAllLocations()
			bresp, err := json.Marshal(locs)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(bresp)
		}
	})

	conn.router.HandleFunc("/api/flow/context/{flowId}", func(w http.ResponseWriter, r *http.Request) {
		if conn.flowContext != nil {
			vars := mux.Vars(r)
			flowId := vars["flowId"]
			var err error
			var bresp []byte
			if flowId == "full_struct_and_states" {
				var result []model2.LocationExtendedView
				states := conn.flowContext.GetDeviceStates()
				locs, _ := conn.assetRegistry.GetAllLocations()
				devs, _ := conn.assetRegistry.GetExtendedDevices()

				for li := range locs {
					rloc := model2.LocationExtendedView{Location:locs[li]}
					for di := range devs {
						if locs[li].ID == devs[di].LocationId {
							for si := range states {
								if states[si].ExternalId == int(devs[di].ID) {
									devs[di].States = append(devs[di].States,states[si])
								}
							}
							rloc.Devices =  append(rloc.Devices, devs[di])
						}

					}
					result = append(result,rloc)
				}
				bresp, err = json.Marshal(result)

			}else {
				records := conn.flowContext.GetRecords(flowId)
				bresp, err = json.Marshal(records)
			}

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Write(bresp)
		}
	})
}

func (conn *Connector) StartHttpServer() {
	conn.server.Handler = conn.router
	if !conn.isServerStarted && conn.config.IsLocalServEnabled {
		log.Info("<HttpConn> Starting HTTP server.")
		go conn.server.ListenAndServe()
		conn.isServerStarted = true
	}
	if !conn.isTunActive && conn.config.IsTunEnabled {
		log.Info("<HttpConn> Starting WS cloud tunnel.")
		conn.StartWsCloudTunnel()
	}
}

// httpFlowRouter is invoked by HTTP server ,and it returns response to caller
func (conn *Connector) httpFlowRouter(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	flowId := vars["id"]
	conn.flowStreamMutex.RLock()
	// Searching by flow id
	stream, ok := conn.flowStreamRegistry[flowId]
	// Searching by alias
	if !ok {
		for i := range conn.flowStreamRegistry {
			if conn.flowStreamRegistry[i].FlowIdAlias == flowId {
				stream = conn.flowStreamRegistry[i]
				ok = true
				break
			}
		}
	}
	conn.flowStreamMutex.RUnlock()

	if code := conn.isRequestAllowed(r, stream.authConfig, flowId); code != AuthCodeAuthorized {
		log.Debug("<HttpConn> Request is not allowed ", flowId)
		conn.SendHttpAuthFailureResponse(code, w, flowId)
		return
	}

	log.Debug("<HttpConn> New HTTP request for flow ", flowId)
	if !ok {
		log.Debug("<HttpConn> no path for ", flowId)
		return
	}

	var responseSignal chan bool
	responseSignal = make(chan bool)
	reqId := utils.GenerateRandomNumber64()
	defer conn.liveConnections.Delete(reqId)
	conn.liveConnections.Store(reqId, liveConnection{respWriter: w, startTime: time.Now(), responseSignal: responseSignal})

	if stream.reqChannel != nil {
		stream.reqChannel <- RequestEvent{RequestId: reqId, HttpRequest: r}
	}

	if responseSignal != nil {
		<-responseSignal
	}

	log.Debug("<HttpConn> http transaction completed. ")
}

// wsFlowRouter is invoked by HTTP server to convert HTTP request to WS request , it start message reading loop , consumes messages from WS stream and
// and routes them to corresponding flow.Method is blocked until connection is alive.

func (conn *Connector) wsFlowRouter(w http.ResponseWriter, r *http.Request) {
	reqId := utils.GenerateRandomNumber()
	defer func() {
		if r := recover(); r != nil {
			log.Error("<HttpConn> WS connection failed with PANIC")
			log.Error(string(debug.Stack()))
		}
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

	if code := conn.isRequestAllowed(r, stream.authConfig, flowId); code != AuthCodeAuthorized {
		log.Debug("<HttpConn> Request is not allowed ", flowId)
		conn.SendHttpAuthFailureResponse(code, w, flowId)
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

	defer func() {
		conn.liveConnections.Delete(reqId)
		log.Debug("<HttpConn> WS connection closed")
	}()

	conn.liveConnections.Store(reqId, liveConnection{respWriter: w, startTime: time.Now(), isWs: true, wsConn: ws, flowId: flowId})

	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Error("<HttpConn> WS Read error :", err)
			break
		} else if msgType == websocket.TextMessage {
			if stream.isPublishOnly {
				continue
			} else {
				if stream.reqChannel != nil {
					stream.reqChannel <- RequestEvent{RequestId: 0, HttpRequest: r, IsWsMsg: true, Payload: msg}
				}
			}
		} else if msgType == websocket.BinaryMessage {
			log.Debug("<HttpConn> New binary ws message")
		} else {
			log.Debug("<HttpConn> Message of type = ", msgType)
		}
	}
}

// RegisterFlow registers node of the flow
func (conn *Connector) RegisterFlow(flowId string, isSync bool, isWs bool, publishOnly bool, reqChannel chan RequestEvent, alias, name string, authConfig AuthConfig) {
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
		FlowIdAlias:   alias,
		Name:          name,
		authConfig:    authConfig,
	}
	conn.flowStreamMutex.Unlock()
	log.Debug("<HttpConn> Registered flow with id = ", flowId)
}

func (conn *Connector) UnregisterFlow(flowId string) {
	conn.liveConnections.Range(func(key, value interface{}) bool {
		// republishing messages to all connected clients
		lConn, ok := value.(liveConnection)
		if !ok {
			return true
		}
		if lConn.flowId == flowId && lConn.isWs {
			lConn.wsConn.Close()
			conn.liveConnections.Delete(key)
			log.Debug("<HttpConn> Connection deleted , id=", key)
		}

		return true
	})

	conn.flowStreamMutex.Lock()
	delete(conn.flowStreamRegistry, flowId)
	conn.flowStreamMutex.Unlock()

	if len(conn.flowStreamRegistry) == 0 {
		conn.server.Close()
		conn.isServerStarted = false
		log.Debug("<HttpConn> No HTTP flows left , shutting down HTTP server")
	}
}

func (conn *Connector) ReplyToRequest(requestId int64, payload []byte, responseContentType string) {
	if requestId == 0 {
		return
	}
	wi, ok := conn.liveConnections.Load(requestId)
	if !ok {
		return
	}
	lConn, ok := wi.(liveConnection)
	if !ok {
		return
	}
	//
	if lConn.isFromCloud {
		// Request was received over cloud channel , sending back to cloud
		headers := map[string]*tunframe.TunnelFrame_StringArray{}
		hVal := tunframe.TunnelFrame_StringArray{Items: []string{responseContentType}}
		headers["Content-Type"] = &hVal
		newMsg := tunframe.TunnelFrame{
			MsgType: tunframe.TunnelFrame_HTTP_RESP,
			Headers: headers,
			CorrId:  requestId,
			Payload: payload,
		}
		conn.tunClient.Send(&newMsg)
		conn.liveConnections.Delete(requestId)
	} else {
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
func (conn *Connector) PublishWs(flowId string, payload []byte) {
	conn.liveConnections.Range(func(key, value interface{}) bool {
		// republishing messages to all connected clients
		lConn, ok := value.(liveConnection)
		if !ok {
			return true
		}
		if lConn.flowId == flowId {
			// TODO : Research is the operation must be executed in async way to avoid blocking , which can happen if client consumes with different speed.
			lConn.wsConn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			err := lConn.wsConn.WriteMessage(websocket.TextMessage, payload)
			if err == nil {
				log.Debug("<httpConn> Message forwarded to client")
			} else {
				log.Info("<httpClient> Can't write to WS connection. Err:", err.Error())
			}
		}
		return true
	})
	// Forward all WS message if cloud connection is active.
	// TODO: ADD SUBSCRIBE support , so cloud has to send subscribe frame to start receiving events
	if conn.isTunActive {
		newMsg := tunframe.TunnelFrame{
			MsgType: tunframe.TunnelFrame_WS_MSG,
			Payload: payload,
		}
		conn.tunClient.Send(&newMsg)
	}

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
