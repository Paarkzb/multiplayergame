package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
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

const worldWidth = int32(1366)
const worldHeight = int32(768)
const minWidth = int32(200)
const minHeight = int32(200)

// Create a Game instance used to handle WebSocket Connections
var game = NewGame()

// Players is a map used to help manage a map of players
type Players map[string]*Player
type Bullets map[string]*Bullet

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := new(sync.WaitGroup)

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

	bot := NewPlayer(nil, "BOT", &Position{X: 250, Y: 250}, 0)
	game.addPlayer(bot)

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Millisecond * 16)

		prevUpdate := time.Now()
		for {
			select {
			case <-ticker.C:
				dt := float64(time.Since(prevUpdate).Milliseconds()) / 1000
				prevUpdate = time.Now()

				player.update(dt)

				for _, bullet := range game.Bullets {

					if bullet.Position.X > float64(worldWidth) || bullet.Position.X < 0 || bullet.Position.Y > float64(worldHeight) || bullet.Position.Y < 0 {
						game.deleteBullet(bullet)
					}

					bullet.update(dt)
				}

				game.checkCollisions(player)

				log.Println("write state in gourutine")
				game.writeState(player)
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
	wg.Wait()
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
		if p.ID != player.ID {
			state.OtherPlayers = append(state.OtherPlayers, p)
		}
	}
	state.Bullets = make([]*Bullet, 0)
	for _, b := range game.Bullets {
		state.Bullets = append(state.Bullets, b)
	}

	log.Println("players ", len(game.Players))

	if player.Conn != nil {
		err := player.Conn.WriteJSON(state)
		if err != nil {
			log.Println(err)
		}
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

func (g *Game) checkCollisions(player *Player) {
	// check players collision
	for _, p := range game.Players {
		if p.ID != player.ID {
			if g.checkPlayerWithPlayerCollision(player, p) {
				// player.canMove = false
				player.Position.X += math.Cos(player.Angle+math.Pi) * 4
				player.Position.Y += math.Sin(player.Angle+math.Pi) * 4
			}
		}
	}
}

func (g *Game) checkPlayerWithPlayerCollision(p1 *Player, p2 *Player) bool {
	return (p1.Position.X+p1.Width >= p2.Position.X &&
		p1.Position.X <= p2.Position.X+p2.Width &&
		p1.Position.Y+p1.Height >= p2.Position.Y &&
		p1.Position.Y <= p2.Position.Y+p2.Height)

}

// addPlayer will add new players to Players
func (g *Game) addPlayer(player *Player) {
	// Lock so we can manipulate
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	// Add Player
	// TODO: Change to normal unique id
	g.Players[player.ID] = player
}

func (g *Game) deletePlayer(player *Player) {
	g.RWMutex.Lock()
	defer g.RWMutex.Unlock()

	log.Println("Deleting player", player.ID)
	delete(g.Players, player.ID)
}

func (g *Game) addBullet(bullet *Bullet) {
	g.Bullets[bullet.ID] = bullet
}

func (g *Game) deleteBullet(bullet *Bullet) {
	delete(g.Bullets, bullet.ID)
}
