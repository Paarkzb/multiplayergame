package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// Event is the Message sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending in the type
type EventHandler func(event Event, player *Player) error

const (
	// EventSendMessage is the event name for messages sent
	EventSendMessage = "send_message"
)

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// SendMessageHandler handle send_message event
func SendMessageHandler(event Event, p *Player) error {
	// Marshal Payload into wanted format
	var messageEvent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &messageEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	log.Println(string(event.Payload))

	return nil
}
