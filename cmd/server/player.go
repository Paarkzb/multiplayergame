package main

import "github.com/gorilla/websocket"

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// Player is a websocket player
type Player struct {
	Id   int32    `json:"id"`
	Name string   `json:"name"`
	Pos  Position `json:"position"`
}

func NewPlayer(id int32, name string, pos Position) *Player {
	return &Player{
		Id:   id,
		Name: name,
		Pos:  pos,
	}
}

type PlayerConn struct {
	Conn   *websocket.Conn
	Player *Player
}

func NewPlayerConn(ws *websocket.Conn, player *Player) *PlayerConn {
	playerConn := &PlayerConn{ws, player}
	// go playerConn.readMessage()
	return playerConn
}

// Players is a map used to help manage a map of players
type PlayerList map[int32]*PlayerConn
