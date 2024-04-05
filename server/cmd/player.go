package main

import (
	"math"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Keys struct {
	A     bool
	D     bool
	W     bool
	S     bool
	Space bool
}

func setKeys() *Keys {
	return &Keys{
		A:     false,
		D:     false,
		W:     false,
		S:     false,
		Space: false,
	}
}

// Player is a websocket player
type Player struct {
	Conn             *websocket.Conn `json:"-"`
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Position         *Position       `json:"position"`
	PreviousPosition *Position       `json:"-"`
	Angle            float64         `json:"angle"`
	Width            float64         `json:"width"`
	Height           float64         `json:"height"`
	Speed            float64         `json:"-"`
	RotateSpeed      float64         `json:"-"`
	Cooldown         float64         `json:"-"`
	Kyes             *Keys           `json:"-"`
	Alive            bool            `json:"-"`
}

func NewPlayer(conn *websocket.Conn, name string, pos *Position, angle float64) *Player {
	return &Player{
		Conn:             conn,
		ID:               uuid.New().String(),
		Name:             name,
		Position:         pos,
		PreviousPosition: &Position{X: pos.X, Y: pos.Y},
		Angle:            angle,
		Width:            50,
		Height:           50,
		Speed:            250,
		RotateSpeed:      125,
		Cooldown:         1,
		Kyes:             setKeys(),
		Alive:            true,
	}
}

func (p *Player) update(dt float64) {
	p.Cooldown += dt
	// log.Println(p.Alive)

	if p.Alive {
		p.PreviousPosition = &Position{X: p.Position.X, Y: p.Position.Y}
		if p.Kyes.A {
			p.Angle -= p.RotateSpeed * math.Pi / 180 * dt
		}
		if p.Kyes.D {
			p.Angle += p.RotateSpeed * math.Pi / 180 * dt
		}
		if p.Kyes.W {
			p.Position.X += math.Cos(p.Angle) * p.Speed * dt
			p.Position.Y += math.Sin(p.Angle) * p.Speed * dt
		}
		if p.Kyes.S {
			p.Position.X -= math.Cos(p.Angle) * p.Speed * dt
			p.Position.Y -= math.Sin(p.Angle) * p.Speed * dt
		}
	}
	if p.Kyes.Space {
		if p.Cooldown > 0.5 {
			p.shoot()
			p.Cooldown = 0
		}
	}

	p.Position.X = math.Min(math.Max(0, p.Position.X), float64(worldWidth-int32(p.Width)))
	p.Position.Y = math.Min(math.Max(0, p.Position.Y), float64(worldHeight-int32(p.Height)))
}

func (p *Player) shoot() {
	cx := p.Position.X + p.Width/2
	cy := p.Position.Y + p.Height/2
	x := p.Position.X + p.Width + 5
	y := p.Position.Y + p.Height/2
	cos := math.Cos(p.Angle)
	sin := math.Sin(p.Angle)
	nx := (cos * (x - cx)) - (sin * (y - cy)) + cx
	ny := (cos * (y - cy)) + (sin * (x - cx)) + cy

	game.addBullet(NewBullet(&Position{
		nx,
		ny,
	}, p.Angle, "common"))
}

func (p *Player) setDead() {
	game.RWMutex.Lock()
	defer game.RWMutex.Unlock()

	p.Alive = false
}
