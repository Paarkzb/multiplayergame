package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

/*
*
websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection
*/
var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Grab the request origin
		// origin := r.Header.Get("Origin")

		// switch origin {
		// case "http://localhost:8080":
		// 	return true
		// default:
		// 	return false
		// }
		return true
	},
}

var ErrEventNotSupported = errors.New("this event type is not supported")

// Manager is used to hold references to all Players Registered, and Broadcasting etc
type Manager struct {
	Players PlayerList

	// Using a syncMutex here to be able to lock state before editing players
	// Could also use the Channels to block
	sync.RWMutex
	// handlers are functions that are used to handle Events
	handlers map[string]EventHandler
}

// NewManager is used to initalize all the values inside the Manager
func NewManager() *Manager {
	m := &Manager{
		Players:  make(PlayerList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

// loginHandler is used to verify an user authentication
func (m *Manager) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "access-control-request-methods,authorization,content-type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST")

	type userLoginRequest struct {
		Username string `json:"username"`
	}

	var req userLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	

	// Failure to auth
	w.WriteHeader(http.StatusUnauthorized)
}

// setupEventHandlers configures and adds all handlers
func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
}

// routeEvent is used to make sure the correct event goes into the correct handler
func (m *Manager) routeEvent(event Event, p *Player) error {
	// Check if Handler is present in Map
	if handler, ok := m.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, p); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
	}
}

// serveWS is a HTTP Handler that the has the Manager that allows connections
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New connection")
	// Begin by upgrading the HTTP request
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create New Player
	player := NewPlayer(conn, m, "Paark")
	m.addPlayer(player)

	log.Println(m.Players)

	// Start the real write processes
	go player.readMessages()

	go player.writeMessage()
}

// addPlayer will add new players to Players
func (m *Manager) addPlayer(player *Player) {
	// Lock so we can manipulate
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()

	// Add Player
	m.Players[player.ID] = player
}

// removePlayer will remove the player and clean up
func (m *Manager) removePlayer(player *Player) {
	// Lock so we can manipulate
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()

	// Check if Player exists, then delete it
	if _, ok := m.Players[player.ID]; ok {
		// close connection
		player.Connection.Close()
		// remove
		delete(m.Players, player.ID)
	}
}
