package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
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

	// EventNewMessage is a response to send_message
	EventNewMessage = "new_message"
)

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

// NewMessageEvent is returned when responding to send_message
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
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

// // SendMessageHandler will send out a message to all players
// func SendMessageToAllPlayers(event Event, p *Player) error {
// 	// Marshal Payload into wanted format
// 	var messageEvent SendMessageEvent
// 	if err := json.Unmarshal(event.Payload, &messageEvent); err != nil {
// 		return fmt.Errorf("bad payload in request: %v", err)
// 	}

// 	// Prepare an Outgoing Message to players
// 	var broadMessage NewMessageEvent

// 	broadMessage.Sent = time.Now()
// 	broadMessage.Message = messageEvent.Message
// 	broadMessage.From = messageEvent.From

// 	data, err := json.Marshal(broadMessage)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal broadcast message: %v", err)
// 	}

// 	// Place payload into an Event
// 	var outgoingEvent Event
// 	outgoingEvent.Payload = data
// 	outgoingEvent.Type = EventNewMessage
// 	// Broadcast to all Players
// 	for _, client := range p.Game.Players {
// 		client.Egress <- outgoingEvent
// 	}

// 	return nil
// }
