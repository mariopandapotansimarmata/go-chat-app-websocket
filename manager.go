package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	ClientList ClientList
	sync.RWMutex

	Handlers map[string]EventHandler
}

func NewManager() *Manager {
	m := &Manager{ClientList: make(ClientList),
		Handlers: make(map[string]EventHandler)}
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
