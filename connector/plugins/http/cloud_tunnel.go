package http

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/edge"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"net/http"
	url2 "net/url"
	"regexp"
	"strings"
	"time"
)

// StartWsCloudTunnel connects to CloudTunnel .
func (conn *Connector) StartWsCloudTunnel() error {
	// 1. Connect to cloud WS endpoint . Request must contain tpflow instance guid and instance token as parameter.
	// 2. Start message reader loop. All messages will be sent over single WS connection.
	// 3. Read message , unpack , route either to WS stream or to Rest stream.
	// 4. Send Rest response back over WS
	log.Info("<HttpCloudConn> Starting cloud tunnel")
	conn.tunClient = edge.NewTunClient(conn.config.TunCloudEndpoint,conn.config.TunAddress,5)
	conn.tunClient.SetEdgeToken(conn.config.TunEdgeToken)

	if err := conn.tunClient.Connect(); err != nil {
		log.Errorf("<HttCloudConn> Connect() error = %v, ", err)
		conn.isTunActive = false
		return err
	}else {
		go conn.wsCloudRouter()
		conn.isTunActive = true
		log.Info("<HttpCloudConn> Successfully connected to ",conn.config.TunCloudEndpoint)
	}
	return nil
}

func (conn *Connector) StopWsCloudTunnel() error {
	log.Info("<HttpCloudConn> Stopping cloud tunnel ")
	conn.tunClient.Close()
	return nil
}
// wsCloudFlowRouter reads messages from cloud stream and routes them either directly to flow or to API.
func (conn *Connector) wsCloudRouter()  {
	stream := conn.tunClient.GetStream()
	for {
		newMsg :=<- stream
		vars := newMsg.GetVars()
		if vars == nil {
			log.Debug("<HttCloudConn> corrupted request, metadata is missing")
			continue
		}
		if vars["tunId"] != conn.config.TunAddress {
			log.Debug("<HttCloudConn> Incorrect tunnel address")
			continue
		}

		log.Info("<HttCloudConn> New msg from cloud stream")
		//log.Debug("%+v",newMsg)
		switch newMsg.GetMsgType() {
		case tunframe.TunnelFrame_HTTP_REQ:
			flowId := vars["flowId"]
			if flowId != "" {
				log.Debug("<HttpCloudConn> New frame for flow ",flowId)
				//if strings.Contains(newMsg.GetReqUrl(),fmt.Sprintf("/flow/%s/rest",flowId)) {
				if r,_ := regexp.MatchString("/cloud/.*/flow/.*/rest.*",newMsg.GetReqUrl());r==true {
					conn.flowStreamMutex.RLock()
					stream, ok := conn.flowStreamRegistry[flowId]
					if !ok {
						for i := range conn.flowStreamRegistry {
							if conn.flowStreamRegistry[i].FlowIdAlias == flowId {
								ok = true
								stream = conn.flowStreamRegistry[i]
							}
						}
					}
					conn.flowStreamMutex.RUnlock()
					if !ok {
						log.Warning("<HttpConn> Flow is not registered. FlowId = ", flowId)
						continue
					}

					httpReq := TunFrameToHttpReq(newMsg)

					if code := conn.isRequestAllowed(httpReq,stream.authConfig,flowId);code != AuthCodeAuthorized {
						log.Debug("<HttpConn> Cloud Request is not allowed ",flowId)
						conn.SendCloudAuthFailureResponse(code,newMsg.ReqId,flowId)
						continue
					}

					if stream.reqChannel != nil {
						conn.liveConnections.Store(newMsg.ReqId, liveConnection{startTime: time.Now(),isFromCloud: true})
						stream.reqChannel <- RequestEvent{RequestId: newMsg.ReqId,Payload: newMsg.Payload,IsFromCloud: true,HttpRequest: httpReq}
					}

				}else if strings.Contains(newMsg.GetReqUrl(),"/api/flow/context/") {
					log.Debug("<HttpCloudConn> New API frame ")
					conn.InternalApiHandlerOverCloud(newMsg)
				}else {
					log.Debug("<HttpCloudConn> Unsupported resource ",newMsg.GetReqUrl())
				}

			}else {
				log.Debug("<HttpCloudConn> New API frame ")
				conn.InternalApiHandlerOverCloud(newMsg)
			}
		case tunframe.TunnelFrame_WS_MSG:
			flowId := vars["flowId"]
			log.Debug("<HttpCloudConn> New WS frame for flow ",flowId)
			if flowId != "" {
				conn.flowStreamMutex.RLock()
				stream, ok := conn.flowStreamRegistry[flowId]
				conn.flowStreamMutex.RUnlock()
				if !ok {
					log.Warning("<HttpConn> Flow is not registered. FlowId = ", flowId)
					continue
				}
				if stream.isPublishOnly {
					continue
				} else {
					if stream.reqChannel != nil {
						stream.reqChannel <- RequestEvent{RequestId: 0, IsWsMsg: true,IsFromCloud: true, Payload: newMsg.Payload}
					}
				}

			}else {
				log.Debug("<HttpCloudConn> Unsupported API ")
			}

		}

	}
	conn.isTunActive = false
}

func (conn *Connector) InternalApiHandlerOverCloud(msg *tunframe.TunnelFrame) {
	var bresp []byte
	var err error
	if strings.Contains(msg.GetReqUrl(),"/api/registry/devices") {
		if conn.assetRegistry != nil {
			devs , _ := conn.assetRegistry.GetExtendedDevices()
			bresp , err =  json.Marshal(devs)
		}
	}else if strings.Contains(msg.GetReqUrl(),"/api/registry/locations") {
		if conn.assetRegistry != nil {
			locs, _ := conn.assetRegistry.GetAllLocations()
			bresp , err =  json.Marshal(locs)
		}
	}else if strings.Contains(msg.GetReqUrl(),"/api/flow/context/") {
		if conn.flowContext != nil {
			flowId := msg.GetVars()["flowId"]
			bresp,err = conn.getContextResponse(flowId)
		}
	}else {
		log.Info("<HttpCloudConn> Unsupported internal API call path ",msg.GetReqUrl())
		return
	}

	tunFrame := tunframe.TunnelFrame{
		MsgType:   tunframe.TunnelFrame_HTTP_RESP,
		CorrId:    msg.ReqId,
		RespCode:  200,
	}
	if err != nil {
		tunFrame.RespCode = http.StatusBadRequest
	}else {
		tunFrame.Payload = bresp
	}
	conn.tunClient.Send(&tunFrame)
}

func (conn *Connector) SendCloudAuthFailureResponse(authCode int , reqId int64,flowId string) {
	tunFrame := tunframe.TunnelFrame{
		MsgType:   tunframe.TunnelFrame_HTTP_RESP,
		Headers:   nil,
		CorrId:    reqId,
		RespCode:  401,
	}

	switch authCode {
	case AuthCodeBasicFailed :
		t := strings.Split(flowId,"_")
		realm := t[0]
		tunFrame.Headers = map[string]*tunframe.TunnelFrame_StringArray{}

		hVal := tunframe.TunnelFrame_StringArray{Items:[]string{fmt.Sprintf(`Basic realm="%s"`,realm)}}
		tunFrame.Headers["WWW-Authenticate"] = &hVal
		tunFrame.Payload = []byte("Unauthorised.\n")
	case AuthCodeFailed:
		//w.WriteHeader(401)
	}
	conn.tunClient.Send(&tunFrame)
}

func TunFrameToHttpReq(tunFrame *tunframe.TunnelFrame) *http.Request {
	headers := http.Header{}
	if tunFrame.Headers != nil {
		for k,v := range tunFrame.Headers {
			headers[k] = v.Items
		}
	}
	r := http.Request{Method: tunFrame.ReqMethod,Header: headers}
	url,err := url2.Parse(tunFrame.ReqUrl)
	if err == nil {
		r.URL = url
	}
	r.RequestURI = tunFrame.ReqUrl
	return &r
}
