package main

import (
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
}

func NewManager() *Manager {
	return &Manager{ClientList: make(ClientList)}
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
