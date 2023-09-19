package executorws

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/rueidis"
)

var upgrader = websocket.Upgrader{}

// Initialize server with empty subscription
var server = NewServer()
var config []utils.ChainConfig

const (
	// time to read the next client's pong message
	pongWait = 60 * time.Second
	// time period to send pings to client
	pingPeriod = (pongWait * 9) / 10
	// time allowed to write a message to client
	writeWait = 10 * time.Second
	// max message size allowed
	maxMessageSize = 512
)

func StartWSServer(config_ []utils.ChainConfig, WS_ADDR string, REDIS_ADDR string, REDIS_PWD string) error {
	config = config_
	client, err := rueidis.NewClient(
		rueidis.ClientOption{InitAddress: []string{REDIS_ADDR}, Password: REDIS_PWD})
	if err != nil {
		return err
	}
	server.RedisClient = &utils.RueidisClient{
		Client: &client,
		Ctx:    context.Background(),
	}
	err = server.RedisClient.Subscribe(utils.CHANNEL_NEW_ORDER, server.handleNewOrder)
	if err != nil {
		return err
	}
	http.HandleFunc("/ws", HandleWs)
	return nil
}

func HandleWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Info("upgrade:" + err.Error())
		return
	}
	defer c.Close()
	// create new client id
	clientID := uuid.New().String()

	//log new client
	slog.Info("Server: new client connected, ID is " + clientID)

	// create channel to signal client health
	done := make(chan struct{})

	go writePump(c, clientID, done)
	readPump(c, clientID, done)
}

// readPump process incoming messages and set the settings
func readPump(conn *websocket.Conn, clientID string, done chan<- struct{}) {
	// set limit, deadline to read & pong handler
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// message handling
	for {
		// read incoming message
		_, msg, err := conn.ReadMessage()
		// if error occured
		if err != nil {
			// remove from the client
			server.RemoveClient(clientID)
			// set health status to unhealthy by closing channel
			close(done)
			// stop process
			break
		}

		// if no error, process incoming message
		server.HandleRequest(conn, clientID, msg)
	}
}

// writePump sends ping to the client
func writePump(conn *websocket.Conn, clientID string, done <-chan struct{}) {
	// create ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// send ping message
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				// if error sending ping, remove this client from the server
				server.RemoveClient(clientID)
				// stop sending ping
				return
			}
		case <-done:
			// if process is done, stop sending ping
			return
		}
	}
}
