package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type State struct {
	Timestamp    int64     `json:"timestamp"`
	Type         string    `json:"type"`
	Player       *Player   `json:"player"`
	OtherPlayers []*Player `json:"otherPlayers"`
	Bullets      []*Bullet `json:"bullets"`
}

// Create a Game instance used to handle WebSocket Connections
var game = NewGame()
var id int32 = 0

// Players is a map used to help manage a map of players
type Players map[int32]*Player
type Bullets []*Bullet

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
	Players Players
	Bullets Bullets
	// Using a syncMutex here to be able to lock state before editing players
	// Could also use the Channels to block
	sync.RWMutex
}

// NewGame is used to initalize all the values inside the Game
func NewGame() *Game {
	g := &Game{
		Players: make(Players),
		Bullets: make(Bullets, 0),
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

	ctx, cancel := context.WithCancel(context.Background())

	player := NewPlayer(conn, id, "", Position{X: 250, Y: 250}, 0)
	game.addPlayer(player)
	id++

	go func() {
		prevUpdate := time.Now()
		for range time.Tick(16 * time.Millisecond) {
			select {
			case <-ctx.Done():
				return
			default:
				dt := float32(time.Since(prevUpdate).Milliseconds()) / 1000
				prevUpdate = time.Now()

				player.update(dt)

				for _, bullet := range game.Bullets {
					bullet.update(dt)
				}

				log.Println("write state in gourutine")
				game.writeState(player)
			}
		}
	}()

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
	player = nil
	cancel()
}

func (g *Game) writeState(player *Player) {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	var state State

	state.Timestamp = time.Now().UnixMilli()
	state.Type = "update"
	state.Player = player
	state.OtherPlayers = make([]*Player, 0)
	for _, p := range game.Players {
		if p.Id != player.Id {
			state.OtherPlayers = append(state.OtherPlayers, p)
		}
	}
	state.Bullets = g.Bullets
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

	case "keydown":
		direction := event.Payload
		// log.Println("Key pressed", event.Payload)
		switch direction {
		case "left":
			player.keys.A = true
		case "right":
			player.keys.D = true
		case "forward":
			player.keys.W = true
		case "back":
			player.keys.S = true
		case "space":
			player.keys.Space = true
		}
	case "keyup":
		direction := event.Payload
		// log.Println("Key unpressed", event.Payload)
		switch direction {
		case "left":
			player.keys.A = false
		case "right":
			player.keys.D = false
		case "forward":
			player.keys.W = false
		case "back":
			player.keys.S = false
		case "space":
			player.keys.Space = false
		}
	}

	// g.writeState(player)
}

// addPlayer will add new players to Players
func (g *Game) addPlayer(player *Player) {
	// Lock so we can manipulate
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	// Add Player
	// TODO: Change to normal unique id
	g.Players[player.Id] = player
}

func (g *Game) deletePlayer(playerId int32) {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	log.Println("Deleting player", playerId)
	delete(g.Players, playerId)
}

func (g *Game) addBullet(bullet *Bullet) {
	g.Bullets = append(g.Bullets, bullet)
}
