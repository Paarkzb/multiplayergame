package main

import "log"

type Setting struct {
	Type       string `json:"type"`
	GameWidth  int32  `json:"game_width"`
	GameHeight int32  `json:"game_height"`
}

const worldWidth = int32(1280)
const worldHeight = int32(720)
const minWidth = int32(200)
const minHeight = int32(200)

func (g *Game) writeConfig(player *Player) {
	var config Setting

	config.Type = "config"
	config.GameWidth = worldWidth
	config.GameHeight = worldHeight

	err := player.Conn.WriteJSON(config)
	if err != nil {
		log.Println(err)
	}
}
