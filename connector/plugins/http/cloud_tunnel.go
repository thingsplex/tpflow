package http



type CloudTunnel struct {

}

// StartWsCloudTunnel connects to cloud tunnel which allows accessing HTTP triggers using public endpoint.
func (conn *CloudTunnel) StartWsCloudTunnel() {
	// 1. Connect to cloud WS endpoint . Request must contain tpflow instance guid and instance token as parameter.
	// 2. Start message reader loop. All messages will be sent over single WS connection.
	// 3. Read message , unpack , route either to WS stream or to Rest stream.
	// 4. Send Rest response back over WS

}

