package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// Message represents the structure of messages exchanged
type Message struct {
	Type    string      `json:"type"`    // "subscribe", "unsubscribe", "publish"
	Topic   string      `json:"topic"`   // The topic name
	Payload interface{} `json:"payload"` // The actual message content
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients for each topic
	subscribers map[string]map[*Client]bool

	// Guard access to the subscribers map
	sync.RWMutex
}

// Client represents a connected websocket client
type Client struct {
	conn *websocket.Conn
	hub  *Hub
}

func newHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) subscribe(topic string, client *Client) {
	h.Lock()
	defer h.Unlock()

	if h.subscribers[topic] == nil {
		h.subscribers[topic] = make(map[*Client]bool)
	}
	h.subscribers[topic][client] = true
	log.Printf("Client subscribed to topic: %s\n", topic)
}

func (h *Hub) unsubscribe(topic string, client *Client) {
	h.Lock()
	defer h.Unlock()

	if subscribers, ok := h.subscribers[topic]; ok {
		delete(subscribers, client)
		if len(subscribers) == 0 {
			delete(h.subscribers, topic)
		}
		log.Printf("Client unsubscribed from topic: %s\n", topic)
	}
}

func (h *Hub) publish(topic string, message Message) {
	h.RLock()
	subscribers := h.subscribers[topic]
	h.RUnlock()

	for client := range subscribers {
		err := wsjson.Write(context.Background(), client.conn, message)
		if err != nil {
			log.Printf("Error sending message to client: %v\n", err)
			// Consider removing the client if we can't write to it
			h.unsubscribe(topic, client)
		}
	}
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("Error accepting websocket: %v\n", err)
		return
	}

	client := &Client{
		conn: conn,
		hub:  hub,
	}

	// Handle incoming messages
	for {
		var msg Message
		err := wsjson.Read(r.Context(), conn, &msg)
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			break
		}

		switch msg.Type {
		case "subscribe":
			hub.subscribe(msg.Topic, client)
			// Send confirmation
			wsjson.Write(r.Context(), conn, Message{
				Type:    "confirmation",
				Topic:   msg.Topic,
				Payload: "Subscribed successfully",
			})

		case "unsubscribe":
			hub.unsubscribe(msg.Topic, client)
			// Send confirmation
			wsjson.Write(r.Context(), conn, Message{
				Type:    "confirmation",
				Topic:   msg.Topic,
				Payload: "Unsubscribed successfully",
			})

		case "publish":
			hub.publish(msg.Topic, msg)
			// Send confirmation
			wsjson.Write(r.Context(), conn, Message{
				Type:    "confirmation",
				Topic:   msg.Topic,
				Payload: "Message published successfully",
			})

		default:
			log.Printf("Unknown message type: %s\n", msg.Type)
		}
	}

	// Cleanup when the connection is closed
	for topic := range hub.subscribers {
		hub.unsubscribe(topic, client)
	}
}

func main() {
	hub := newHub()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	log.Println("WebSocket server starting on :2021")
	log.Fatal(http.ListenAndServe(":2021", nil))
}
