package main

func (g *Game) checkCollisions(player *Player) {
	// check player with player collision
	game.Players.Range(func(_, op interface{}) bool {
		p := op.(*Player)

		if p.ID != player.ID {
			if g.checkPlayerWithPlayerCollision(player, p) {
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
				g.writeMessage("end", player)
				// g.deletePlayer(p)
			}

			return true
		})

		return true
	})
}

func (g *Game) checkPlayerWithPlayerCollision(p1 *Player, p2 *Player) bool {
	return (p1.Position.X+p1.Width >= p2.Position.X &&
		p1.Position.X <= p2.Position.X+p2.Width &&
		p1.Position.Y+p1.Height >= p2.Position.Y &&
		p1.Position.Y <= p2.Position.Y+p2.Height)

}

func (g *Game) checkBulletWithPlayerCollision(b *Bullet, p *Player) bool {
	closestX := clamp(b.Position.X, p.Position.X, p.Position.X+p.Width)
	closestY := clamp(b.Position.Y, p.Position.Y, p.Position.Y+p.Height)

	distanceX := b.Position.X - closestX
	distanceY := b.Position.Y - closestY

	distanceSquared := (distanceX * distanceX) + (distanceY * distanceY)

	return distanceSquared < (b.Radius * b.Radius)
}
