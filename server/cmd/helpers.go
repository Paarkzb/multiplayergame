package main

import (
	"log"
	"math"
)

// addPlayer will add new players to Players
func (g *Game) addPlayer(player *Player) {
	g.Players.Store(player.ID, player)
}

func (g *Game) deletePlayer(player *Player) {
	log.Println("Deleting player", player.ID)
	g.Players.Delete(player.ID)
}

func (g *Game) addBullet(bullet *Bullet) {
	g.Bullets.Store(bullet.ID, bullet)
}

func (g *Game) deleteBullet(bullet *Bullet) {
	g.Bullets.Delete(bullet.ID)
}

func clamp(val float64, lo float64, hi float64) float64 {
	return math.Max(lo, math.Min(val, hi))
}

func (g *Game) updateBullets(dt float64) {
	game.Bullets.Range(func(_, ob interface{}) bool {
		b := ob.(*Bullet)

		b.update(dt)

		if b.Position.X > float64(worldWidth) || b.Position.X < 0 || b.Position.Y > float64(worldHeight) || b.Position.Y < 0 {
			game.deleteBullet(b)
		}

		return true
	})
}
