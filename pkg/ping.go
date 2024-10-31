package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pingInterval    = time.Second
	httpTimeout     = 5 * time.Second
	dataBytes       = 56
	responseBytes   = 64
	defaultTTL      = 50
	websocketBuffer = 1024
)

// PingMessage represents the incoming ping request
type PingMessage struct {
	Address string `json:"address"` // The address to ping (IP or domain)
}

// PongMessage represents the ping response with latency information
type PongMessage struct {
	Type      string    `json:"type"`      // Message type ("pong")
	Timestamp time.Time `json:"timestamp"` // Time when the response was created
	Bytes     int       `json:"bytes"`     // Number of bytes in the response
	Sequence  int       `json:"sequence"`  // Sequence number of the ping
	Address   string    `json:"address"`   // Address that was pinged
	Latency   float64   `json:"latency"`   // Round-trip time in milliseconds
	Success   bool      `json:"success"`   // Whether the ping was successful
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  websocketBuffer,
	WriteBufferSize: websocketBuffer,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// formatAddress ensures the address has the correct protocol prefix
func formatAddress(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return "http://" + addr
	}
	return addr
}

// measureLatency performs the HTTP GET request and measures the round-trip time
func measureLatency(client *http.Client, address string) (float64, error) {
	startTime := time.Now()
	resp, err := client.Get(address)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return float64(time.Since(startTime).Microseconds()) / 1000.0, nil
}

// createPongMessage creates a PongMessage with the given parameters
func createPongMessage(address string, sequence int, latency float64, success bool) PongMessage {
	return PongMessage{
		Type:      "pong",
		Timestamp: time.Now(),
		Bytes:     responseBytes,
		Sequence:  sequence,
		Address:   address,
		Latency:   latency,
		Success:   false,
	}
}

// sendPongMessage sends the pong message through the websocket connection
func sendPongMessage(conn *websocket.Conn, msg PongMessage) error {
	if err := conn.WriteJSON(msg); err != nil {
		return fmt.Errorf("error writing pong: %w", err)
	}
	return nil
}

// checkConnection verifies if the websocket connection is still alive
func checkConnection(conn *websocket.Conn) error {
	deadline := time.Now().Add(pingInterval)
	if err := conn.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
		return fmt.Errorf("client disconnected: %w", err)
	}
	return nil
}

// logPingResult logs the ping result in the standard ping format
func logPingResult(address string, sequence int, latency float64, success bool) {
	if !success {
		log.Printf("Request timeout for icmp_seq=%d", sequence)
		return
	}

	log.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms",
		responseBytes,
		address,
		sequence,
		defaultTTL,
		latency,
	)
}

// PingHandler handles WebSocket ping requests
func PingHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Read initial ping message
	var ping PingMessage
	if err := conn.ReadJSON(&ping); err != nil {
		log.Printf("Error reading ping message: %v", err)
		return
	}

	// Format address and initialize
	ping.Address = formatAddress(ping.Address)
	log.Printf("PING %s (%s): %d data bytes", ping.Address, ping.Address, dataBytes)

	// Initialize ping client and ticker
	client := &http.Client{Timeout: httpTimeout}
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	sequence := 0

	// Main ping loop
	for range ticker.C {
		sequence++

		// Measure latency
		latency, err := measureLatency(client, ping.Address)
		success := err == nil

		// Create and send pong message
		pong := createPongMessage(ping.Address, sequence-1, latency, success)
		if err := sendPongMessage(conn, pong); err != nil {
			log.Printf("Failed to send pong: %v", err)
			return
		}

		// Log result
		logPingResult(ping.Address, sequence-1, latency, success)

		// Check connection health
		if err := checkConnection(conn); err != nil {
			log.Printf("Connection check failed: %v", err)
			return
		}
	}
}
