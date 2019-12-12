package gwebsocket

import (
	ws "github.com/gorilla/websocket"
	gc "github.com/jholowczak/guacamole_client_go"
	gn "github.com/jholowczak/guacamole_client_go/gnet"
	gp "github.com/jholowczak/guacamole_client_go/gprotocol"
	"strconv"
	"sync"
	"time"
)

const (
	BUFFER_SIZE int    = 8192
	PING_OPCODE string = "ping"
)

type GuacamoleWebSocketTunnelEndpoint struct {
	tunnel *gn.GuacamoleTunnel
	conn   *ws.Conn
	mu     *sync.Mutex
}

func CreateTunnelEndpoint(conn *ws.Conn, socket *gn.GuacamoleTunnel) (ret *GuacamoleWebSocketTunnelEndpoint) {
	ret = &GuacamoleWebSocketTunnelEndpoint{
		tunnel: socket,
		conn:   conn,
		mu:     &sync.Mutex{},
	}
	return
}

type GuacamoleWebSocketTunnelEndpointInterface interface {
	closeConnection(int, int) error
	closeConnectionWithStatus(gc.GuacamoleStatus) error
	sendInstruction(string) error
	sendGuacamoleInstruction(gp.GuacamoleInstruction) error
	//OnOpen(*ws.Conn)
	OnMessage(message string)
	//OnClose(*ws.Conn)
	Tunnel() gn.GuacamoleTunnel
}

func (te *GuacamoleWebSocketTunnelEndpoint) CloseConnection(guacStatusCode, webSocketCode int) (err error) {
	te.mu.Lock()
	defer te.mu.Unlock()
	err = te.conn.WriteControl(webSocketCode, []byte(strconv.Itoa(guacStatusCode)), time.Now().Add(time.Second*5))
	return
}

func (te *GuacamoleWebSocketTunnelEndpoint) CloseConnectionWithStatus(guacStatus gc.GuacamoleStatus) (err error) {
	err = te.CloseConnection(guacStatus.GetGuacamoleStatusCode(), guacStatus.GetWebSocketCode())
	return
}

func (te *GuacamoleWebSocketTunnelEndpoint) SendInstruction(instruction string) (err error) {
	te.mu.Lock()
	defer te.mu.Unlock()
	err = te.conn.WriteMessage(1, []byte(instruction))
	return
}

func (te *GuacamoleWebSocketTunnelEndpoint) SendGuacamoleInstruction(instruction gp.GuacamoleInstruction) (err error) {
	err = te.SendInstruction(instruction.String())
	return
}

func (te *GuacamoleWebSocketTunnelEndpoint) Tunnel() *gn.GuacamoleTunnel {
	return te.tunnel
}

func (te GuacamoleWebSocketTunnelEndpoint) Filter(instruction gp.GuacamoleInstruction) (ret gp.GuacamoleInstruction, exc gc.ExceptionInterface) {
	if instruction.GetOpcode() == gn.InternalDataOpcode {
		args := instruction.GetArgs()
		if len(args) >= 2 && args[0] == PING_OPCODE {
			err := te.SendGuacamoleInstruction(gp.NewGuacamoleInstruction(gn.InternalDataOpcode, PING_OPCODE, args[1]))
			if err != nil {
				exc = gc.GuacamoleException.Throw("Unable to send ping response for websocket tunnel")
			}
			return
		}
	}
	ret = instruction
	return
}
