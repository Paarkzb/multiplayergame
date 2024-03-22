package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Create a Game instance used to handle WebSocket Connections
var game = NewGame()

/*
*
websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection
*/
var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Grab the request origin
		// origin := r.Header.Get("Origin")

		// switch origin {
		// case "http://localhost:8080":
		// 	return true
		// default:
		// 	return false
		// }
		return true
	},
}

var ErrEventNotSupported = errors.New("this event type is not supported")

// Game is used to hold references to all Players Registered, and Broadcasting etc
type Game struct {
	Players PlayerList

	// Using a syncMutex here to be able to lock state before editing players
	// Could also use the Channels to block
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers map[string]EventHandler
}

// NewGame is used to initalize all the values inside the Game
func NewGame() *Game {
	g := &Game{
		Players:  make(PlayerList),
		handlers: make(map[string]EventHandler),
	}
	g.setupEventHandlers()
	return g
}

// loginHandler is used to verify an user authentication
func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")

	var username struct {
		Username string `json:"username"`
	}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&username)
	if err != nil {
		log.Println(err)
	}

	player := NewPlayer(0, username.Username, Position{X: 0, Y: 0})
	game.addPlayer(nil, player)

	resp, err := json.MarshalIndent(player, "", "\t")
	if err != nil {
		log.Println(err)
	}

	log.Println("Logged")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// serveWS is a HTTP Handler that the has the Game that allows connections
func serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New connection")
	// Begin by upgrading the HTTP request
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// handle incoming messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// handle message
		var event Event
		err = json.Unmarshal(message, &event)
		if err != nil {
			log.Println(err)
			continue
		}

		game.handleMessages(event, game)

	}

	// Send game data to clients after connect
	// go game.sendGameDataToClient(conn)

	// // Start the real write processes
	// go player.readMessages()

	// go player.writeMessage()
}

// Send game data to clients
func (g *Game) sendGameDataToClient(conn *websocket.Conn) {
	data, _ := json.Marshal(g.Players)

	err := conn.WriteJSON(Event{Type: EventSendMessage, Payload: data})
	log.Println("send data")
	if err != nil {
		log.Println(err)
	}
}

// Response to the client after getting message
func (g *Game) broadcast(event Event) {
	for _, player := range g.Players {
		err := player.Conn.WriteJSON(event)
		if err != nil {
			log.Println(err)
		}
	}
}

func (g *Game) handleMessages(event Event, game *Game) {

	// handle any msg type
	var payload struct {
		Message Player
		From    int32
	}
	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Println(err)
	}

	switch event.Type {
	case "login":
		log.Println(payload)
		game.Players[payload.From].Player.Name = payload.Message.Name
	case "move":
		game.broadcast(event)
	case "action":
		log.Println("ACTION")
	}
}

// addPlayer will add new players to Players
func (g *Game) addPlayer(conn *websocket.Conn, player *Player) {
	// Lock so we can manipulate
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	// Add Player
	// TODO: Change to normal unique id
	playerConn := NewPlayerConn(conn, player)
	g.Players[player.Id] = playerConn
}

// // removePlayer will remove the player and clean up
// func (g *Game) removePlayer(player *Player) {
// 	// Lock so we can manipulate
// 	g.RWMutex.Lock()
// 	defer g.RWMutex.Unlock()

// 	// Check if Player exists, then delete it
// 	if _, ok := g.Players[player.ID]; ok {
// 		// close connection
// 		player.Connection.Close()
// 		// remove
// 		delete(g.Players, player.ID)
// 	}
// }

// setupEventHandlers configures and adds all handlers
func (g *Game) setupEventHandlers() {
	g.handlers[EventSendMessage] = SendMessageHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (g *Game) routeEvent(event Event, p *Player) error {
	// Check if Handler is present in Map
	if handler, ok := g.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, p); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}
