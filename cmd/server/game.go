package main

import (
	"encoding/json"
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

// Game is used to hold references to all Players Registered, and Broadcasting etc
type Game struct {
	Players PlayerList

	// // Using a syncMutex here to be able to lock state before editing players
	// // Could also use the Channels to block
	// sync.RWMutex

}

// NewGame is used to initalize all the values inside the Game
func NewGame() *Game {
	g := &Game{
		Players: make(PlayerList),
	}

	return g
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

	player := NewPlayer(conn, id, "", Position{X: 0, Y: 0})
	game.addPlayer(player)
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

		game.handleMessages(event, player)

	}
	game.deletePlayer(player.Id)
}

// Response to the client after getting message
// func (g *Game) broadcast(event Event) {
// 	for _, player := range g.Players {
// 		err := player.Conn.WriteJSON(event)
// 		if err != nil {
// 			log.Println(err, "broadcast")
// 		}
// 	}
// }

func (g *Game) writeState(player *Player) {
	var state struct {
		Type         string   `json:"type"`
		Player       Player   `json:"player"`
		OtherPlayers []Player `json:otherPlayers`
	}
	state.Type = "update"
	state.Player = *player
	for _, p := range game.Players {
		if p.Id != player.Id {
			state.OtherPlayers = append(state.OtherPlayers, *p)
		}
	}
	log.Println("players ", len(game.Players))

	err := player.Conn.WriteJSON(state)
	if err != nil {
		log.Println(err)
	}
}

func (g *Game) handleMessages(event Event, player *Player) {
	switch event.Type {
	case "login":
		player.Name = event.Payload

	case "move":
		direction := event.Payload
		switch direction {
		case "left":
			player.Pos.X -= 10
		case "right":
			player.Pos.X += 10
		case "up":
			player.Pos.Y -= 10
		case "down":
			player.Pos.Y += 10
		}
	}

	g.writeState(player)
}

// addPlayer will add new players to Players
func (g *Game) addPlayer(player *Player) {
	// // Lock so we can manipulate
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	// Add Player
	// TODO: Change to normal unique id
	g.Players[player.Id] = player
}

func (g *Game) deletePlayer(playerId int32) {
	log.Println("Deleting player", playerId)
	delete(g.Players, playerId)
}
