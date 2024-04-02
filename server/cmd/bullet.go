package main

import "math"

type Bullet struct {
	Position   Position `json:"position"`
	Angle      float32  `json:"angle"`
	BulletType string   `json:"bullet_type"`
	Speed      float32  `json:"speed"`
}

func NewBullet(position Position, angle float32, bulletType string) *Bullet {
	return &Bullet{Position: position, Angle: angle, BulletType: bulletType, Speed: 50}
}

func (b *Bullet) update(dt float32) {
	rad := float64(b.Angle)
	b.Position.X += float32(math.Cos(rad)) * b.Speed * dt
	b.Position.Y += float32(math.Sin(rad)) * b.Speed * dt
}
