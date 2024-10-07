package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	egress chan Event
}

func NewClient(conn *websocket.Conn, m *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    m,
		egress:     make(chan Event),
	}
}

func (c *Client) ReadMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	for {
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("erro reading message : %v", err)
			}
			break
		}
		request := Event{}

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marhalling event: %v", err)
			break
		}
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("err handling message: ", err)
		}
	}
}

func (c *Client) WriteMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed", err)
				}
			}
			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("failed to send message: %v", err)
			}
			log.Println("message sent")
		}
	}
}
