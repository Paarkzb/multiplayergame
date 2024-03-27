package main

import "github.com/gorilla/websocket"

type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type Keys struct {
	A bool
	D bool
	W bool
	S bool
}

// Player is a websocket player
type Player struct {
	Conn  *websocket.Conn `json:"-"`
	Id    int32           `json:"id"`
	Name  string          `json:"name"`
	Pos   Position        `json:"position"`
	Angle float32         `json:"angle"`
	Speed float32         `json:"-"`
	keys  Keys            `json:"-"`
}

func NewPlayer(conn *websocket.Conn, id int32, name string, pos Position, angle float32) *Player {
	return &Player{
		Conn:  conn,
		Id:    id,
		Name:  name,
		Pos:   pos,
		Angle: angle,
		Speed: 10,
		keys:  Keys{A: false, D: false, W: false, S: false},
	}
}

// Players is a map used to help manage a map of players
type PlayerList map[int32]*Player
