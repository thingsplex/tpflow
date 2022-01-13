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
	reconnectCounter int
	readErrorCounter int
	connHeaders      http.Header // Should be used to configure additional metadata , like security tokens
	sendQueue        chan []byte
}

func NewTunClient(host string, edgeConnId string, streamBufferSize int) *TunClient {
	return &TunClient{host: host, edgeConnId: edgeConnId, streamBufferSize: streamBufferSize, connHeaders: http.Header{}}
}

func (tc *TunClient) SetEdgeToken(token string) {
	tc.connHeaders.Set("X-TPlex-Token", token)
}

func (tc *TunClient) SetHeader(key, value string) {
	tc.connHeaders.Set(key, value)
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
		tc.readErrorCounter = 0
		log.Debug("<edgeClient> Successfully connected")
	} else {
		tc.IsConnected = false
	}
	tc.sendQueue = make(chan []byte, tc.streamBufferSize)
	go tc.startMsgDeliveryWorker()
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
	if binMsg, err := proto.Marshal(msg); err != nil {
		log.Info("<edgeClient> Failed to parse proto message")
		return err
	} else {
		// TODO : introduce write buffer/queue to avoid blocking by slow connection
		return tc.wsConn.WriteMessage(websocket.BinaryMessage, binMsg)
	}
}

// SendAsync sends tunnel frame over open WS connection
func (tc *TunClient) SendAsync(msg *tunframe.TunnelFrame) error {
	//var binMsg []byte
	if !tc.IsConnected {
		return fmt.Errorf("ws client disconnected")
	}
	if binMsg, err := proto.Marshal(msg); err != nil {
		log.Info("<edgeClient> Failed to parse proto message")
		return err
	} else {
		tc.sendQueue <- binMsg
	}
	return nil
}

func (tc *TunClient) startMsgDeliveryWorker() {
	for binMsg := range tc.sendQueue {
		if !tc.IsConnected {
			log.Errorf(" ws client disconnected")
			continue
		}
		tc.wsConn.WriteMessage(websocket.BinaryMessage, binMsg)
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
	close(tc.sendQueue)
	tc.IsConnected = false
}

func (tc *TunClient) startMsgReader() {
	for {
		if !tc.IsRunning {
			break
		}
		if !tc.IsConnected {
			time.Sleep(time.Second * 3)
			tc.reconnectCounter++
			if err := tc.Connect(); err != nil {
				delay := time.Duration(2 * tc.reconnectCounter)
				if delay > 1200 {
					delay = 1200
				}
				log.Warningf("<edgeClient> Reconnection failed...reconnecting after %d sec", delay)
				time.Sleep(time.Second * delay)
			}
			continue
		}

		msgType, msg, err := tc.wsConn.ReadMessage() // reading message from Edge devices
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Warning("<edgeClient> Lost connection , reconnecting after 2 sec")
				tc.IsConnected = false
			} else if websocket.IsCloseError(err) {
				tc.IsConnected = false
				log.Warning("<edgeClient> WS Close error :", err)
			} else {
				log.Warning("<edgeClient> WS Read error :", err)
				if tc.readErrorCounter > 5 {
					tc.wsConn.Close()
					tc.IsConnected = false
					time.Sleep(time.Second * 3)
				}
				tc.readErrorCounter++
			}
			time.Sleep(time.Second * 3)
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
