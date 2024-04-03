package main

import (
	"math"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
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
	Conn        *websocket.Conn `json:"-"`
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Position    Position        `json:"position"`
	Angle       float32         `json:"angle"`
	Width       float32         `json:"width"`
	Speed       float32         `json:"-"`
	RotateSpeed float32         `json:"-"`
	Cooldown    float32         `json:"-"`
	keys        *Keys           `json:"-"`
}

func NewPlayer(conn *websocket.Conn, name string, pos Position, angle float32) *Player {
	return &Player{
		Conn:        conn,
		ID:          uuid.New().String(),
		Name:        name,
		Position:    pos,
		Angle:       angle,
		Width:       50,
		Speed:       250,
		RotateSpeed: 125,
		Cooldown:    0,
		keys:        setKeys(),
	}
}

func (p *Player) update(dt float32) {
	p.Cooldown += dt
	if p.keys.A {
		p.Angle -= p.RotateSpeed * math.Pi / 180 * dt
	}
	if p.keys.D {
		p.Angle += p.RotateSpeed * math.Pi / 180 * dt
	}
	if p.keys.W {
		p.Position.X += float32(math.Cos(float64(p.Angle))) * p.Speed * dt
		p.Position.Y += float32(math.Sin(float64(p.Angle))) * p.Speed * dt
	}
	if p.keys.S {
		p.Position.X -= float32(math.Cos(float64(p.Angle))) * p.Speed * dt
		p.Position.Y -= float32(math.Sin(float64(p.Angle))) * p.Speed * dt
	}
	if p.keys.Space {
		if p.Cooldown > 1 {
			p.shoot()
			p.Cooldown = 0
		}
	}
}

func (p *Player) shoot() {
	// bullet := NewBullet(p.Position, p.Angle, "common")
	x := p.Position.X + p.Width/2
	y := p.Position.Y + p.Width/2
	x += float32(math.Cos(float64(p.Angle))) * 30
	y += float32(math.Cos(float64(p.Angle))) * 30
	game.addBullet(NewBullet(&Position{
		x,
		y,
	}, p.Angle, "common"))
}
