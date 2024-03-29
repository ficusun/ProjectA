package wsserver

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Client - struct
type Client struct {
	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send   chan []byte
	ID     int
	Status bool
	Remove bool
	Auth   bool
}

func (c *Client) writePump() {

	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.Status = false
		
		mes, err := json.Marshal(Letter{c.ID, "1901", "Off"})
		if err != nil {
			log.Println(err)
		}
		FromConnChan <- mes
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {

	defer func() {
		c.Conn.Close()
		c.Status = false

		mes, err := json.Marshal(Letter{c.ID, "1901", "Off"})
		if err != nil {
			log.Println(err)
		}
		FromConnChan <- mes
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		lettToStr := string(message)
		typeLett := lettToStr[0:4]
		letter := lettToStr[4:]

		if (*c).Auth == false && typeLett != "1001" {

			mes, err := json.Marshal(struct{ Failed string }{"You are not auth"})
			if err != nil {
				log.Println(err)
			}
			c.Send <- mes
			continue
		}

		mes, err := json.Marshal(Letter{c.ID, typeLett, letter})
		if err != nil {
			log.Println(err)
		}
		FromConnChan <- mes
	}
}

// start - func for start methods of client
func (c *Client) start() {

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go c.writePump()
	go c.readPump()
}

// GetID - return Client ID
// type int
func (c *Client) GetID() int {
	return c.ID
}
