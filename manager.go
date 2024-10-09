package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		CheckOrigin:     checkkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	ClientList ClientList
	sync.RWMutex

	otps     RetentionMap
	Handlers map[string]EventHandler
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		ClientList: make(ClientList),
		Handlers:   make(map[string]EventHandler),
		otps:       NewRetentionMap(ctx, 5),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.Handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	fmt.Println(event)
	return nil
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.Handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("there is no such event type")
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	otp := r.URL.Query().Get("otp")
	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !m.otps.VerifyOTP(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Println("New Connection")

	// Upgrade regular http connection into websocket
	conn, err := webSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	client := NewClient(conn, m)

	m.addClient(client)
	go client.ReadMessage()
	go client.WriteMessage()
}

func (m *Manager) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type userLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	var loginReq userLoginRequest
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&loginReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if loginReq.Username == "user" && loginReq.Password == "user" {
		type Response struct {
			OTP string `json:"otp"`
		}
		otp := m.otps.NewOTP()
		resp := Response{
			OTP: otp.Key,
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
}

func (m *Manager) addClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	m.ClientList[c] = true

}

func (m *Manager) removeClient(c *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.ClientList[c]; ok {
		c.connection.Close()
		delete(m.ClientList, c)
	}
}

func checkkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	log.Println(origin)

	switch origin {
	case "https://localhost:8080":
		return true
	default:
		return false
	}
}
