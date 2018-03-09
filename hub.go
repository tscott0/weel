package main

import (
	"log"
	"net/http"
)

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// ServeHTTP handles websocket requests from the peer.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Println("hub: Failed to get session")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	val := session.Values["logged_in"]
	var loggedIn bool
	var ok bool
	if loggedIn, ok = val.(bool); !ok || !loggedIn {
		log.Println("hub: User isn't logged in")
		return
	}

	val = session.Values["username"]
	var username string
	if username, ok = val.(string); !ok {
		log.Println("hub: Can't find username in session")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan Message, 256),
		username: username,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
