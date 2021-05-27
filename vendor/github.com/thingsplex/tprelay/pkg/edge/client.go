package edge

import (
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tprelay/pkg/proto/tunframe"
	"google.golang.org/protobuf/proto"
	"net/http"
	"time"
)

type TunClient struct {
	host             string // ws://localhost:8083
	edgeConnId       string
	wsConn           *websocket.Conn
	stream           chan *tunframe.TunnelFrame
	streamBufferSize int
	stopSignal       chan bool
	IsConnected      bool
	IsRunning        bool
	connHeaders      http.Header // Should be used to configure additional metadata , like security tokens
}

func NewTunClient(host string, edgeConnId string, streamBufferSize int) *TunClient {
	return &TunClient{host: host, edgeConnId: edgeConnId, streamBufferSize: streamBufferSize,connHeaders: http.Header{}}
}

func (tc *TunClient) SetEdgeToken(token string) {
	tc.connHeaders.Set("X-TPlex-Token",token)
}

func (tc *TunClient) SetHeader(key,value string) {
	tc.connHeaders.Set(key,value)
}

func (tc *TunClient) Connect() error {
	// TODO : Introduce connection retry.
	var err error
	tc.stopSignal = make(chan bool)
	address := fmt.Sprintf("%s/edge/%s/register", tc.host, tc.edgeConnId)
	tc.wsConn, _, err = websocket.DefaultDialer.Dial(address, tc.connHeaders)
	tc.IsRunning = true
	if err == nil {
		tc.IsConnected = true
		log.Debug("<edgeClient> Successfully connected")
	}else {
		tc.IsConnected = false
	}
	return err
}

func (tc *TunClient) SetConnHeaders(connHeaders http.Header) {
	tc.connHeaders = connHeaders
}

// Send sends tunnel frame over open WS connection
func (tc *TunClient) Send(msg *tunframe.TunnelFrame) error {
	//var binMsg []byte
	if !tc.IsConnected {
		return fmt.Errorf("client disconnected")
	}
	if binMsg,err := proto.Marshal(msg); err != nil {
		log.Info("<edgeClient> Failed to parse proto message")
		return err
	}else {
		// TODO : introduce write buffer/queue to avoid blocking by slow connection
		return tc.wsConn.WriteMessage(websocket.BinaryMessage, binMsg)
	}
}

// GetStream returns channel that should be used to consume frames from open connection
func (tc *TunClient) GetStream() chan *tunframe.TunnelFrame {
	if tc.stream == nil {
		tc.stream = make(chan *tunframe.TunnelFrame, tc.streamBufferSize)
		go tc.startMsgReader()
	}
	return tc.stream
}

func (tc *TunClient) Close() {
	if tc.stream != nil {
		tc.stopSignal <- true
	}
	tc.wsConn.Close()
	tc.IsConnected = false
}


func (tc *TunClient) startMsgReader() {
	for {
		if !tc.IsRunning {
			break
		}
		if !tc.IsConnected {
			time.Sleep(time.Second*3)
			if err := tc.Connect();err != nil {
				log.Warning("<edgeClient> Reconnection failed..")
			}
			continue
		}
		msgType, msg, err := tc.wsConn.ReadMessage() // reading message from Edge devices
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Warning("<edgeClient> Lost connection , reconnecting after 2 sec")
				tc.IsConnected = false
				time.Sleep(time.Second*2)
			}else {
				log.Warning("<edgeClient> WS Read error :", err)
			}
			continue
		}
		if msgType != websocket.BinaryMessage {
			log.Debug("<edgeClient> Unsupported msg type")
			continue
		}

		tunMsg := tunframe.TunnelFrame{}
		if err := proto.Unmarshal(msg, &tunMsg); err != nil {
			log.Info("<edgeClient> Failed to parse proto message")
			continue
		}

		select {
		case tc.stream <- &tunMsg:
			log.Debug("<edgeClient> Message forwarded")
		case <-tc.stopSignal:
			break
		default:
		}
	}
	log.Debug("<edgeClient> Msg reader stopped")
}

