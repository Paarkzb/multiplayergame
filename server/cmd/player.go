package main

import (
	"math"
	"time"

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
	Id          int32           `json:"id"`
	Name        string          `json:"name"`
	Position    Position        `json:"position"`
	Angle       float32         `json:"angle"`
	Speed       float32         `json:"-"`
	RotateSpeed float32         `json:"-"`
	Cooldown    time.Duration   `json:"-"`
	keys        *Keys           `json:"-"`
}

func NewPlayer(conn *websocket.Conn, id int32, name string, pos Position, angle float32) *Player {
	return &Player{
		Conn:        conn,
		Id:          id,
		Name:        name,
		Position:    pos,
		Angle:       angle,
		Speed:       250,
		RotateSpeed: 125,
		Cooldown:    3 * time.Millisecond,
		keys:        setKeys(),
	}
}

func (p *Player) update(dt float32) {
	if p.keys.A {
		p.Angle -= p.RotateSpeed * math.Pi / 180 * dt
	}
	if p.keys.D {
		p.Angle += p.RotateSpeed * math.Pi / 180 * dt
	}
	if p.keys.W {
		rad := float64(p.Angle)
		p.Position.X += float32(math.Cos(rad)) * p.Speed * dt
		p.Position.Y += float32(math.Sin(rad)) * p.Speed * dt
	}
	if p.keys.S {
		rad := float64(p.Angle)
		p.Position.X -= float32(math.Cos(rad)) * p.Speed * dt
		p.Position.Y -= float32(math.Sin(rad)) * p.Speed * dt
	}
	if p.keys.Space {
		p.shoot()
	}
}

func (p *Player) shoot() {
	bullet := NewBullet(p.Position, p.Angle, "common")

	game.addBullet(bullet)
}
