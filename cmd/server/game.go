package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Create a Game instance used to handle WebSocket Connections
var game = NewGame()
var id int32 = 0

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

	// // Using a syncMutex here to be able to lock state before editing players
	// // Could also use the Channels to block
	// sync.RWMutex
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

	player := NewPlayer(id, username.Username, Position{X: 0, Y: 0})
	game.addPlayer(nil, player)
	id++

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
		log.Fatal(err)
		return
	}
	defer conn.Close()

	player := NewPlayer(id, "", Position{X: 0, Y: 0})
	game.addPlayer(conn, player)
	id++

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

		var payload struct {
			Message json.RawMessage
			From    int32
		}
		err = json.Unmarshal(event.Payload, &payload)
		if err != nil {
			log.Println(err)
		}

		game.Players[payload.From].Conn = conn

		game.handleMessages(event, conn)

	}
	game.deletePlayer(player.Id)

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
			log.Println(err, "broadcast")
		}
	}
}

func (g *Game) handleMessages(event Event, conn *websocket.Conn) {
	// handle any msg type
	var payload struct {
		Message json.RawMessage
		From    int32
	}
	err := json.Unmarshal(event.Payload, &payload)
	if err != nil {
		log.Println(err)
	}

	log.Println("Handling message from ", payload.From)
	player := g.Players[payload.From].Player

	var respEvent Event
	respEvent.Type = event.Type

	switch event.Type {
	case "login":
		var playerPayload Player
		err := json.Unmarshal(payload.Message, &playerPayload)
		if err != nil {
			log.Println(err)
		}
		player.Name = playerPayload.Name

	case "move":
		var direction struct {
			Direction string `json:"direction"`
		}
		log.Println(string(payload.Message))
		err := json.Unmarshal(payload.Message, &direction)
		if err != nil {
			log.Println(err)
		}

		switch direction.Direction {
		case "left":
			player.Pos.X -= 10
		case "right":
			player.Pos.X += 10
		case "up":
			player.Pos.Y -= 10
		case "down":
			player.Pos.Y += 10
		}

	case "close":
		log.Println("ACTION")
	default:
	}

	state.Player = *player
	for _, p := range game.Players {
		if p.Player.Id != player.Id {
			state.OtherPlayers = append(state.OtherPlayers, *p.Player)
		}
	}
	log.Println("players ", len(game.Players))
	out, err := json.Marshal(state)
	if err != nil {
		log.Println(err)
	}
	event.Payload = out

	err = conn.WriteJSON(event)
	if err != nil {
		log.Println(err)
	}
}

// addPlayer will add new players to Players
func (g *Game) addPlayer(conn *websocket.Conn, player *Player) {
	// // Lock so we can manipulate
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	// Add Player
	// TODO: Change to normal unique id
	playerConn := NewPlayerConn(conn, player)
	g.Players[player.Id] = playerConn
}

func (g *Game) deletePlayer(playerId int32) {
	log.Println("Deleting player", playerId)
	delete(g.Players, playerId)
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
