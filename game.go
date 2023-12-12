package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less then pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead * 9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is because otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

// Players is a map used to help manage a map of players
type PlayerList map[int]*Player

// Player is a websocket player, basically a frontend visitor
type Player struct {
	// the websocket connection
	Connection *websocket.Conn

	// Manager is used to manage the players
	Manager *Manager

	// Egress is used to avoid concurrent writes on the WebSocket
	Egress chan Event

	ID   int
	Name string
	X    int
	Y    int
}

func NewPlayer(conn *websocket.Conn, manager *Manager, name string) *Player {
	return &Player{
		Connection: conn,
		Manager:    manager,
		Egress:     make(chan Event),
		ID:         0,
		Name:       name,
		X:          0,
		Y:          0,
	}
}

func (p *Player) PongHandler(pongMsg string) error {
	// Current time + Pong Wait time
	log.Println("pong")
	return p.Connection.SetReadDeadline(time.Now().Add(pongWait))
}

// readMessages will start the player to read messages and handle them
// appropriatly
// This is suppose to be run as a goroutine
func (p *Player) readMessages() {
	defer func() {
		// Close the Connection once this function is done
		p.Manager.removePlayer(p)
	}()

	// Set Max Size of Messages in Bytes
	p.Connection.SetReadLimit(512)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := p.Connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	// Configure how to handle Pong responses
	p.Connection.SetPongHandler(p.PongHandler)

	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := p.Connection.ReadMessage()
		if err != nil {
			// If Connection is closed, we will Receive an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message: %v", err)
			break // Breaking the connection here might be harsh
		}

		// Route the Event
		if err := p.Manager.routeEvent(request, p); err != nil {
			log.Println("Error handeling Message: ", err)
		}
	}
}

// WriteMessage is a process that listens for new messages to output to the Client
func (p *Player) writeMessage() {
	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		// close if this triggers a clossing
		p.Manager.removePlayer(p)
	}()

	for {
		select {
		case message, ok := <-p.Egress:
			// Ok will be false Incase the Egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicates that to frontend
				if err := p.Connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the gorutine
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return // closes the connection, should we really
			}
			log.Println(string(data))
			// Write a Regular text message to the connection
			if err := p.Connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		case <-ticker.C:
			log.Println("ping")
			// Send the Ping
			if err := p.Connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg: ", err)
				return //
			}
		}
	}
}
