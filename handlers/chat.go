package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"RideUP/hub"
	"RideUP/middleware"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type incomingMsg struct {
	Body     string `json:"body"`
	Username string `json:"username"`
}

type outgoingMsg struct {
	SenderID string    `json:"sender_id"`
	Username string    `json:"username"`
	Body     string    `json:"body"`
	SentAt   time.Time `json:"sent_at"`
}

func ChatWS(h *hub.Hub) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("eventId")
		if eventID == "" {
			http.Error(w, `{"error":"missing eventId"}`, http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok || userID == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		token := r.URL.Query().Get("token")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}

		client := &hub.Client{
			EventID: eventID,
			UserID:  userID,
			Conn:    conn,
			Send:    make(chan []byte, 64),
		}
		h.Register <- client

		go writePump(client)
		readPump(client, h, token)
	})
}

func readPump(c *hub.Client, h *hub.Hub, token string) {
	defer func() {
		h.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var in incomingMsg
		if err := json.Unmarshal(raw, &in); err != nil || in.Body == "" {
			continue
		}

		now := time.Now().UTC()

		if err := persistMessage(c.EventID, c.UserID, in.Body, token); err != nil {
			log.Printf("persist message: %v", err)
		}

		out := outgoingMsg{
			SenderID: c.UserID,
			Username: in.Username,
			Body:     in.Body,
			SentAt:   now,
		}
		data, _ := json.Marshal(out)

		h.Broadcast <- &hub.Message{EventID: c.EventID, Data: data}
	}
}

func persistMessage(eventID, userID, body, token string) error {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil
	}

	payload, _ := json.Marshal(map[string]string{
		"event_id": eventID,
		"user_id":  userID,
		"content":  body,
	})

	req, err := http.NewRequest(http.MethodPost,
		supabaseURL+"/rest/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))
	req.Header.Set("Prefer", "return=minimal")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("supabase insert: status %d", resp.StatusCode)
	}
	return nil
}

func writePump(c *hub.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
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
