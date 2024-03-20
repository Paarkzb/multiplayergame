package main

import "github.com/gorilla/websocket"

type Position struct {
	x float32
	y float32
}

// Player is a websocket player, basically a frontend visitor
type Player struct {
	Id   int
	Name string
	Pos  Position
}

func NewPlayer(id int32, name string, pos Position) *Player {
	return &Player{
		Id:   0,
		Name: name,
		Pos:  Position{0, 0},
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
type PlayerList map[int]*PlayerConn
