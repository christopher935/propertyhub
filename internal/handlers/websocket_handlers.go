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
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *WebSocketHub
	UserID   string
	UserRole string
}

type WebSocketHub struct {
	clients      map[*WebSocketClient]bool
	broadcast    chan []byte
	register     chan *WebSocketClient
	unregister   chan *WebSocketClient
	mu           sync.RWMutex
	statsService *services.DashboardStatsService
}

func NewWebSocketHub(statsService *services.DashboardStatsService) *WebSocketHub {
	hub := &WebSocketHub{
		clients:      make(map[*WebSocketClient]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *WebSocketClient),
		unregister:   make(chan *WebSocketClient),
		statsService: statsService,
	}
	go hub.Run()
	if statsService != nil {
		go hub.periodicUpdate()
	}
	return hub
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected: %s (Total: %d)", client.ID, len(h.clients))
			
			if h.statsService != nil {
				h.BroadcastVisitorCount()
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("WebSocket client disconnected: %s (Total: %d)", client.ID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
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
	if h.statsService == nil {
		return
	}
	
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
	
	message := map[string]interface{}{
		"type": "visitor_count",
		"data": map[string]interface{}{
			"count":           activeVisitors,
			"trend":           visitorsTrend,
			"by_page":         visitorsByPage,
			"hot_count":       hotVisitors,
			"returning_count": returningVisitors,
			"timestamp":       time.Now().Unix(),
		},
	}
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling visitor count: %v", err)
		return
	}
	
	h.broadcast <- jsonData
}

func (h *WebSocketHub) BroadcastToAdmins(messageType string, data interface{}) error {
	message := map[string]interface{}{
		"type":      messageType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.UserRole == "admin" || client.UserRole == "super_admin" {
			select {
			case client.Send <- jsonData:
			default:
				log.Printf("Failed to send to client %s", client.ID)
			}
		}
	}

	return nil
}

func (h *WebSocketHub) GetConnectedAdminCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for client := range h.clients {
		if client.UserRole == "admin" || client.UserRole == "super_admin" {
			count++
		}
	}
	return count
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "ping":
				c.Send <- []byte(`{"type":"pong"}`)
			case "mark_read":
				log.Printf("Client %s marked notification as read", c.ID)
			}
		}
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WebSocketHandler struct {
	hub *WebSocketHub
}

func NewWebSocketHandler(db *gorm.DB, statsService *services.DashboardStatsService) *WebSocketHandler {
	hub := NewWebSocketHub(statsService)
	return &WebSocketHandler{hub: hub}
}

func (wsh *WebSocketHandler) HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	userRole := "guest"
	userID := "anonymous"

	if role, exists := c.Get("user_role"); exists {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	}

	if id, exists := c.Get("user_id"); exists {
		if idStr, ok := id.(string); ok {
			userID = idStr
		}
	}

	client := &WebSocketClient{
		ID:       userID + "_" + time.Now().Format("20060102150405"),
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      wsh.hub,
		UserID:   userID,
		UserRole: userRole,
	}

	wsh.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (wsh *WebSocketHandler) GetHub() *WebSocketHub {
	return wsh.hub
}

func (wsh *WebSocketHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"connected_admins": wsh.hub.GetConnectedAdminCount(),
		"total_clients":    len(wsh.hub.clients),
	})
}
