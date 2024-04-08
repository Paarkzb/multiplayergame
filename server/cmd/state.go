package main

import (
	"log"
	"time"
)

type State struct {
	Timestamp    int64     `json:"timestamp"`
	Type         string    `json:"type"`
	Player       *Player   `json:"player"`
	OtherPlayers []*Player `json:"otherPlayers"`
	Bullets      []*Bullet `json:"bullets"`
}

func (g *Game) writeState(player *Player) {
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
