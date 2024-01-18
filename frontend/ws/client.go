package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoorchat/frontend/components"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	// TODO use to check auth periodically
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// NewClient also starts the readPump and writePump.
func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	cl := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go cl.readPump()
	go cl.writePump()

	return cl
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil },
	)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("error: %v", err)
			}
			break
		}
		// devlog.Print("sending message:", string(message))
		c.hub.broadcast <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			//
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			var incoming map[string]json.RawMessage
			err = json.Unmarshal(message, &incoming)
			if err != nil {
				log.Println("failed to Unmarshall message:", string(message))
				continue
			}
			devlog.Printf("received message %#v\n", string(message))

			//
			if val, ok := incoming["text"]; ok {
				text := val[1 : len(val)-1]       // remove quotes
				text = bytes.TrimRight(text, " ") // remove trailing spaces
				components.Message(string(text)).Render(context.TODO(), w)
			}

			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	components.ChatMessageTemplate.ExecuteTemplate(w, "msg", string(<-c.send))
			// }

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
