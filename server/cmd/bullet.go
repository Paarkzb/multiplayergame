package main

import (
	"math"

	"github.com/google/uuid"
)

type Bullet struct {
	ID         string    `json:"id"`
	Position   *Position `json:"position"`
	Angle      float64   `json:"angle"`
	BulletType string    `json:"bullet_type"`
	Speed      float64   `json:"speed"`
}

func NewBullet(position *Position, angle float64, bulletType string) *Bullet {
	return &Bullet{ID: uuid.New().String(), Position: position, Angle: angle, BulletType: bulletType, Speed: 300}
}

func (b *Bullet) update(dt float64) {
	rad := b.Angle
	b.Position.X += math.Cos(rad) * b.Speed * dt
	b.Position.Y += math.Sin(rad) * b.Speed * dt
}
