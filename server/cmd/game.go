package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

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

// Game is used to hold references to all Players Registered, and Broadcasting etc
type Game struct {
	Players sync.Map
	Bullets sync.Map
}

// NewGame is used to initalize all the values inside the Game
func NewGame() *Game {
	g := &Game{}

	return g
}

// serveWS is a HTTP Handler that the has the Game that allows connections
func serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New connection")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Begin by upgrading the HTTP request
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	random := rand.New(rand.NewSource(time.Now().Unix()))
	player := NewPlayer(conn, "", &Position{
		X: float64(random.Int31n(worldWidth-2*minWidth) + minWidth),
		Y: float64(random.Int31n(worldHeight-2*minHeight) + minHeight),
	}, 0)

	game.addPlayer(player)

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Millisecond * 16)

		prevUpdate := time.Now()
		for {
			select {
			case <-ticker.C:
				dt := float64(time.Since(prevUpdate).Milliseconds()) / 1000
				prevUpdate = time.Now()

				player.update(dt)

				game.updateBullets(dt)

				game.checkCollisions(player)

				log.Println("write state in gourutine")
				game.writeMessage("state", player)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}(ctx)

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

	game.deletePlayer(player)
}

func (g *Game) handleMessages(event Event, player *Player) {
	switch event.Type {
	case "login":
		player.Name = event.Payload
		g.writeMessage("start", player)

	case "keydown":
		direction := event.Payload
		// log.Println("Key pressed", event.Payload)
		switch direction {
		case "left":
			player.Kyes.A = true
		case "right":
			player.Kyes.D = true
		case "forward":
			player.Kyes.W = true
		case "back":
			player.Kyes.S = true
		case "space":
			player.Kyes.Space = true
		}
	case "keyup":
		direction := event.Payload
		// log.Println("Key unpressed", event.Payload)
		switch direction {
		case "left":
			player.Kyes.A = false
		case "right":
			player.Kyes.D = false
		case "forward":
			player.Kyes.W = false
		case "back":
			player.Kyes.S = false
		case "space":
			player.Kyes.Space = false
		}
	}

	// g.writeState(player)
}
