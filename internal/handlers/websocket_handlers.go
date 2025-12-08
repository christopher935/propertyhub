package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"chrisgross-ctrl-project/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type WebSocketClient struct {
	conn *websocket.Conn
	send chan WebSocketMessage
	hub  *WebSocketHub
}

type WebSocketHub struct {
	clients      map[*WebSocketClient]bool
	broadcast    chan WebSocketMessage
	register     chan *WebSocketClient
	unregister   chan *WebSocketClient
	mu           sync.RWMutex
	statsService *services.DashboardStatsService
}

type ActivityEvent struct {
	Type       string                 `json:"type"`
	SessionID  string                 `json:"session_id"`
	UserID     int64                  `json:"user_id,omitempty"`
	UserEmail  string                 `json:"user_email,omitempty"`
	PropertyID *int64                 `json:"property_id,omitempty"`
	Details    string                 `json:"details"`
	Timestamp  time.Time              `json:"timestamp"`
	Score      int                    `json:"score,omitempty"`
	EventData  map[string]interface{} `json:"event_data,omitempty"`
}

type ActivityHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan ActivityEvent
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mu         sync.RWMutex
	db         *gorm.DB
}

func NewWebSocketHub(statsService *services.DashboardStatsService) *WebSocketHub {
	hub := &WebSocketHub{
		clients:      make(map[*WebSocketClient]bool),
		broadcast:    make(chan WebSocketMessage, 256),
		register:     make(chan *WebSocketClient),
		unregister:   make(chan *WebSocketClient),
		statsService: statsService,
	}
	go hub.run()
	go hub.periodicUpdate()
	return hub
}

func (h *WebSocketHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered. Total clients: %d", len(h.clients))

			h.BroadcastVisitorCount()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WebSocketHub) periodicUpdate() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.BroadcastVisitorCount()
	}
}

func (h *WebSocketHub) BroadcastVisitorCount() {
	stats, err := h.statsService.GetLiveStats()
	if err != nil {
		log.Printf("Error getting live stats for WebSocket broadcast: %v", err)
		return
	}

	activeVisitors, _ := stats["active_visitors"].(int64)
	visitorsTrend, _ := stats["visitors_trend"].(int64)
	visitorsByPage, _ := stats["visitors_by_page"].(map[string]int)
	hotVisitors, _ := stats["hot_visitors"].(int64)
	returningVisitors, _ := stats["returning_visitors"].(int64)

	h.Broadcast(WebSocketMessage{
		Type: "visitor_count",
		Data: map[string]interface{}{
			"count":           activeVisitors,
			"trend":           visitorsTrend,
			"by_page":         visitorsByPage,
			"hot_count":       hotVisitors,
			"returning_count": returningVisitors,
			"timestamp":       time.Now().Unix(),
		},
	})
}

func (h *WebSocketHub) Broadcast(message WebSocketMessage) {
	h.broadcast <- message
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling WebSocket message: %v", err)
				return
			}
			w.Write(data)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WebSocketHandler struct {
	hub         *WebSocketHub
	activityHub *ActivityHub
}

func NewWebSocketHandler(db *gorm.DB, statsService *services.DashboardStatsService) *WebSocketHandler {
	hub := NewWebSocketHub(statsService)
	activityHub := NewActivityHub(db)
	return &WebSocketHandler{
		hub:         hub,
		activityHub: activityHub,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		conn: conn,
		send: make(chan WebSocketMessage, 256),
		hub:  h.hub,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *WebSocketHandler) GetHub() *WebSocketHub {
	return h.hub
}

func NewActivityHub(db *gorm.DB) *ActivityHub {
	hub := &ActivityHub{
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan ActivityEvent, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		db:         db,
	}
	go hub.run()
	return hub
}

func (h *ActivityHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Activity client registered. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Activity client unregistered. Total clients: %d", len(h.clients))

		case event := <-h.broadcast:
			h.mu.RLock()
			message := WebSocketMessage{
				Type: "activity_event",
				Data: map[string]interface{}{
					"event": event,
				},
			}
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *ActivityHub) BroadcastEvent(event ActivityEvent) {
	h.broadcast <- event
}

func (h *ActivityHub) BroadcastActiveCount(count int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := WebSocketMessage{
		Type: "active_count",
		Data: map[string]interface{}{
			"count":     count,
			"timestamp": time.Now().Unix(),
		},
	}

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *WebSocketHandler) HandleAdminActivityFeed(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		conn: conn,
		send: make(chan WebSocketMessage, 256),
		hub:  h.hub,
	}

	h.activityHub.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *WebSocketHandler) GetActivityHub() *ActivityHub {
	return h.activityHub
}
