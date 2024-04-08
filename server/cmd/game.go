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

type Config struct {
	Type       string `json:"type"`
	GameWidth  int32  `json:"game_width"`
	GameHeight int32  `json:"game_height"`
}

const worldWidth = int32(1280)
const worldHeight = int32(720)
const minWidth = int32(200)
const minHeight = int32(200)

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
	// Using a syncMutex here to be able to lock state before editing players
	sync.RWMutex

	sync.WaitGroup
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
	// game.WaitGroup.Wait()
}

func (g *Game) writeState(player *Player) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	var state State

	state.Timestamp = time.Now().UnixMilli()
	state.Type = "update"
	state.Player = player

	state.OtherPlayers = make([]*Player, 0)
	game.Players.Range(func(_, op interface{}) bool {
		p := op.(*Player)

		state.OtherPlayers = append(state.OtherPlayers, p)

		return true
	})

	state.Bullets = make([]*Bullet, 0)
	game.Bullets.Range(func(_, ob interface{}) bool {
		b := ob.(*Bullet)

		state.Bullets = append(state.Bullets, b)

		return true
	})

	playerLength := 0
	game.Players.Range(func(_, _ interface{}) bool {
		playerLength++

		return true
	})
	log.Println("players ", playerLength)

	if player.Conn != nil {
		err := player.Conn.WriteJSON(state)
		if err != nil {
			log.Println(err)
		}
	}
}

func (g *Game) writeConfig(player *Player) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	var config Config

	config.Type = "config"
	config.GameWidth = worldWidth
	config.GameHeight = worldHeight

	err := player.Conn.WriteJSON(config)
	if err != nil {
		log.Println(err)
	}
}

func (g *Game) handleMessages(event Event, player *Player) {
	switch event.Type {
	case "login":
		player.Name = event.Payload
		g.writeConfig(player)

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

func (g *Game) updateBullets(dt float64) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	game.Bullets.Range(func(_, ob interface{}) bool {
		b := ob.(*Bullet)

		b.update(dt)

		if b.Position.X > float64(worldWidth) || b.Position.X < 0 || b.Position.Y > float64(worldHeight) || b.Position.Y < 0 {
			game.deleteBullet(b)
		}

		return true
	})
}

func (g *Game) checkCollisions(player *Player) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	// check player with player collision
	game.Players.Range(func(_, op interface{}) bool {
		p := op.(*Player)

		if p.ID != player.ID {
			if g.checkPlayerWithPlayerCollision(player, p) {
				// player.canMove = false
				// player.Position.X += math.Cos(player.Angle+math.Pi) * 4
				// player.Position.Y += math.Sin(player.Angle+math.Pi) * 4

				player.Position = player.PreviousPosition
			}
		}

		return true
	})

	// check bullet with player collision
	game.Bullets.Range(func(_, ob interface{}) bool {
		b := ob.(*Bullet)

		game.Players.Range(func(_, op interface{}) bool {
			p := op.(*Player)

			if g.checkBulletWithPlayerCollision(b, p) {
				g.deleteBullet(b)
				p.setDead()
				// g.deletePlayer(p)
			}

			return true
		})

		return true
	})
}

func (g *Game) checkPlayerWithPlayerCollision(p1 *Player, p2 *Player) bool {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	return (p1.Position.X+p1.Width >= p2.Position.X &&
		p1.Position.X <= p2.Position.X+p2.Width &&
		p1.Position.Y+p1.Height >= p2.Position.Y &&
		p1.Position.Y <= p2.Position.Y+p2.Height)

}

func clamp(val float64, lo float64, hi float64) float64 {
	return math.Max(lo, math.Min(val, hi))
}

func (g *Game) checkBulletWithPlayerCollision(b *Bullet, p *Player) bool {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	closestX := clamp(b.Position.X, p.Position.X, p.Position.X+p.Width)
	closestY := clamp(b.Position.Y, p.Position.Y, p.Position.Y+p.Height)

	distanceX := b.Position.X - closestX
	distanceY := b.Position.Y - closestY

	distanceSquared := (distanceX * distanceX) + (distanceY * distanceY)

	return distanceSquared < (b.Radius * b.Radius)
}

// addPlayer will add new players to Players
func (g *Game) addPlayer(player *Player) {
	// Lock so we can manipulate
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	g.Players.Store(player.ID, player)
}

func (g *Game) deletePlayer(player *Player) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	log.Println("Deleting player", player.ID)
	g.Players.Delete(player.ID)
}

func (g *Game) addBullet(bullet *Bullet) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	g.Bullets.Store(bullet.ID, bullet)
}

func (g *Game) deleteBullet(bullet *Bullet) {
	// g.RWMutex.Lock()
	// defer g.RWMutex.Unlock()

	// delete(g.Bullets, bullet.ID)
	g.Bullets.Delete(bullet.ID)
}
