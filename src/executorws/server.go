package executorws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"log/slog"

	"github.com/D8-X/d8x-broker-server/src/utils"
	"github.com/gorilla/websocket"
)

// Subscriptions is a type for each string of topic and the clients that subscribe to it
type Subscriptions map[string]Clients

// Clients is a type that describe the clients' ID and their connection
type Clients map[string]*websocket.Conn

// Server is the struct to handle the Server functions & manage the Subscriptions
type Server struct {
	Subscriptions Subscriptions
	RedisClient   *utils.RueidisClient
}

type ClientMessage struct {
	Type  string `json:"type"`
	Topic string `json:"topic"`
}

type ServerResponse struct {
	Type  string      `json:"type"`
	Topic string      `json:"topic"`
	Data  interface{} `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const NEW_ORDER_TOPIC = "orders"

func NewServer() *Server {
	var s Server
	s.Subscriptions = make(Subscriptions)
	s.Subscriptions[NEW_ORDER_TOPIC] = make(Clients)
	return &s
}

// Send simply sends message to the websocket client
func (s *Server) Send(conn *websocket.Conn, message []byte) {
	// send simple message
	conn.WriteMessage(websocket.TextMessage, message)
}

// SendWithWait sends message to the websocket client using wait group, allowing usage with goroutines
func (s *Server) SendWithWait(conn *websocket.Conn, message []byte, wg *sync.WaitGroup) {
	// send simple message
	conn.WriteMessage(websocket.TextMessage, message)

	// set the task as done
	wg.Done()
}

// RemoveClient removes the clients from the server subscription map
func (s *Server) RemoveClient(clientID string) {
	// loop all topics
	for _, client := range s.Subscriptions {
		// delete the client from all the topic's client map
		delete(client, clientID)
	}
}

// Process incoming websocket message
// https://github.com/madeindra/golang-websocket/
func (s *Server) HandleRequest(conn *websocket.Conn, clientID string, message []byte) {
	slog.Info("recv: " + fmt.Sprint(message))
	var data ClientMessage
	err := json.Unmarshal(message, &data)
	if err != nil {
		// JSON parsing not successful
		return
	}
	reqTopic := strings.TrimSpace(strings.ToLower(data.Topic))
	reqType := strings.TrimSpace(strings.ToLower(data.Type))
	if reqType == "subscribe" {
		msg := s.SubscribeOrders(conn, clientID, reqTopic)
		server.Send(conn, msg)
	} else if reqType == "unsubscribe" {
		// unsubscribe
		s.UnsubscribeOrders(clientID, reqTopic)
	} //else: ignore
}

func (s *Server) AckSub(topic string) []byte {
	r := ServerResponse{Type: "subscribe", Topic: topic, Data: "ack"}
	jsonData, err := json.Marshal(r)
	if err != nil {
		slog.Error("forming response")
		return []byte{}
	}
	return jsonData
}

// Subscribe the client to new orders
func (s *Server) SubscribeOrders(conn *websocket.Conn, clientID string, topic string) []byte {
	if !isValidOrderTopic(topic) {
		return errorResponse("subscribe", topic, "usage: perpetualId:chainId")
	}
	if _, exist := s.Subscriptions[topic]; exist {
		clients := s.Subscriptions[topic]
		// if client already subscribed, stop the process
		if _, subbed := clients[clientID]; subbed {
			return errorResponse("subscribe", topic, "client already subscribed")
		}
		// not subscribed
		clients[clientID] = conn
		return s.AckSub(topic)
	}
	// if topic does not exist, create a new topic
	newClients := make(Clients)
	s.Subscriptions[topic] = newClients

	// add the client to the topic
	s.Subscriptions[topic][clientID] = conn
	return s.AckSub(topic)
}

// Unsubscribe the client from a candle-topic (e.g. btc-usd:15m)
func (s *Server) UnsubscribeOrders(clientID string, topic string) {
	// if topic exists, check the client map
	if _, exist := s.Subscriptions[topic]; exist {
		client := s.Subscriptions[topic]
		// remove the client from the topic's client map
		delete(client, clientID)
	}
}

func isValidOrderTopic(topic string) bool {
	pattern := "^[0-9]+:[0-9]+$" // Regular expression for order topics
	regex, _ := regexp.Compile(pattern)
	if !regex.MatchString(topic) {
		return false
	}
	perpId, chainId, _ := strings.Cut(topic, ":")
	id, _ := strconv.Atoi(perpId)
	if id < 100000 {
		return false
	}
	// supported chainId?
	id, _ = strconv.Atoi(chainId)
	for _, el := range config {
		if el.ChainId == int64(id) {
			return true
		}
	}
	return false
}

func errorResponse(reqType string, reqTopic string, msg string) []byte {

	e := ErrorResponse{Error: msg}
	res := ServerResponse{Type: reqType, Topic: reqTopic, Data: e}
	jsonData, err := json.Marshal(res)
	if err != nil {
		slog.Error("forming error response")
	}
	return jsonData
}
