package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Default values for ping options
const (
	defaultCount      = 0    // 0 means ping continuously
	defaultWait       = 1    // 1 second between pings
	defaultTTL        = 64   // Default TTL value
	defaultPacketSize = 56   // Default packet size in bytes
	defaultTimeout    = 5    // 5 second timeout
	defaultWaitTime   = 10   // 10 second wait time
	defaultTOS        = 0    // Type of Service
	defaultPreload    = 0    // Number of packets to preload
	defaultSweepMin   = 0    // Minimum sweep size
	defaultSweepMax   = 0    // Maximum sweep size
	defaultSweepIncr  = 0    // Sweep increment size
	websocketBuffer   = 1024 // WebSocket buffer size
)

// PingMessage represents the incoming ping request with optional fields
type PingMessage struct {
	// Required
	Address string `json:"address"` // The address to ping (IP or domain)

	// Optional flags
	Adaptive  *bool `json:"a,omitempty"`      // Adaptive ping (-A)
	Audible   *bool `json:"a_flag,omitempty"` // Audible ping (-a)
	Debug     *bool `json:"d,omitempty"`      // Debug mode (-d)
	Flood     *bool `json:"f,omitempty"`      // Flood ping (-f)
	Numeric   *bool `json:"n,omitempty"`      // Numeric output only (-n)
	Quiet     *bool `json:"q,omitempty"`      // Quiet output (-q)
	Timestamp *bool `json:"D,omitempty"`      // Print timestamp (-D)
	Verbose   *bool `json:"v,omitempty"`      // Verbose output (-v)

	// Optional parameters with values
	Count         *int    `json:"c,omitempty"` // Number of pings to send (-c)
	SweepMaxSize  *int    `json:"G,omitempty"` // Maximum sweep size (-G)
	SweepMinSize  *int    `json:"g,omitempty"` // Minimum sweep size (-g)
	SweepIncrSize *int    `json:"h,omitempty"` // Sweep increment size (-h)
	Wait          *int    `json:"i,omitempty"` // Interval between pings (-i)
	Preload       *int    `json:"l,omitempty"` // Number of packets to preload (-l)
	Mask          *string `json:"M,omitempty"` // Mask or time (-M)
	TTL           *int    `json:"m,omitempty"` // Time to live (-m)
	Pattern       *string `json:"p,omitempty"` // Pattern to fill packets (-p)
	SourceAddr    *string `json:"S,omitempty"` // Source address (-S)
	PacketSize    *int    `json:"s,omitempty"` // Packet size (-s)
	Timeout       *int    `json:"t,omitempty"` // Timeout (-t)
	WaitTime      *int    `json:"W,omitempty"` // Wait time for responses (-W)
	TOS           *int    `json:"z,omitempty"` // Type of Service (-z)
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

// PingOptions contains the resolved ping options
type PingOptions struct {
	Count         int
	Wait          int
	TTL           int
	PacketSize    int
	Timeout       int
	WaitTime      int
	TOS           int
	Preload       int
	SweepMinSize  int
	SweepMaxSize  int
	SweepIncrSize int
	SourceAddr    string
	Pattern       string
	Mask          string
	IsAdaptive    bool
	IsAudible     bool
	IsDebug       bool
	IsFlood       bool
	IsNumeric     bool
	IsQuiet       bool
	HasTimestamp  bool
	IsVerbose     bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  websocketBuffer,
	WriteBufferSize: websocketBuffer,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// getOrDefault is a helper function to handle optional values
func getOrDefault[T any](value *T, defaultValue T) T {
	if value == nil {
		return defaultValue
	}
	return *value
}

// validatePingOptions validates and adjusts ping options if needed
func validatePingOptions(opts *PingOptions) error {
	if opts.Count < 0 {
		return fmt.Errorf("count cannot be negative")
	}
	if opts.Wait < 0 {
		return fmt.Errorf("wait interval cannot be negative")
	}
	if opts.TTL <= 0 || opts.TTL > 255 {
		return fmt.Errorf("TTL must be between 1 and 255")
	}
	if opts.PacketSize < 0 {
		return fmt.Errorf("packet size cannot be negative")
	}
	if opts.SweepMaxSize > 0 {
		if opts.SweepMinSize >= opts.SweepMaxSize {
			return fmt.Errorf("sweep min size must be less than max size")
		}
		if opts.SweepIncrSize <= 0 {
			return fmt.Errorf("sweep increment size must be positive")
		}
	}
	if opts.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if opts.WaitTime <= 0 {
		return fmt.Errorf("wait time must be positive")
	}
	if opts.TOS < 0 || opts.TOS > 255 {
		return fmt.Errorf("TOS must be between 0 and 255")
	}
	return nil
}

// resolvePingOptions converts PingMessage to PingOptions with defaults
func resolvePingOptions(msg *PingMessage) (PingOptions, error) {
	opts := PingOptions{
		Count:         getOrDefault(msg.Count, defaultCount),
		Wait:          getOrDefault(msg.Wait, defaultWait),
		TTL:           getOrDefault(msg.TTL, defaultTTL),
		PacketSize:    getOrDefault(msg.PacketSize, defaultPacketSize),
		Timeout:       getOrDefault(msg.Timeout, defaultTimeout),
		WaitTime:      getOrDefault(msg.WaitTime, defaultWaitTime),
		TOS:           getOrDefault(msg.TOS, defaultTOS),
		SweepMinSize:  getOrDefault(msg.SweepMinSize, defaultSweepMin),
		SweepMaxSize:  getOrDefault(msg.SweepMaxSize, defaultSweepMax),
		SweepIncrSize: getOrDefault(msg.SweepIncrSize, defaultSweepIncr),
		Preload:       getOrDefault(msg.Preload, defaultPreload),
		SourceAddr:    getOrDefault(msg.SourceAddr, ""),
		Pattern:       getOrDefault(msg.Pattern, ""),
		Mask:          getOrDefault(msg.Mask, ""),
		IsAdaptive:    getOrDefault(msg.Adaptive, false),
		IsAudible:     getOrDefault(msg.Audible, false),
		IsDebug:       getOrDefault(msg.Debug, false),
		IsFlood:       getOrDefault(msg.Flood, false),
		IsNumeric:     getOrDefault(msg.Numeric, false),
		IsQuiet:       getOrDefault(msg.Quiet, false),
		HasTimestamp:  getOrDefault(msg.Timestamp, false),
		IsVerbose:     getOrDefault(msg.Verbose, false),
	}

	if err := validatePingOptions(&opts); err != nil {
		return opts, fmt.Errorf("invalid ping options: %w", err)
	}

	if opts.IsFlood {
		opts.Wait = 1
	}

	return opts, nil
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
		Bytes:     defaultPacketSize,
		Sequence:  sequence,
		Address:   address,
		Latency:   latency,
		Success:   success,
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
	deadline := time.Now().Add(time.Second)
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
		defaultPacketSize,
		address,
		sequence,
		defaultTTL,
		latency,
	)
}

// PingHandler handles WebSocket ping requests
func PingHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	var pingMsg PingMessage
	if err := conn.ReadJSON(&pingMsg); err != nil {
		log.Printf("Error reading ping message: %v", err)
		return
	}

	opts, err := resolvePingOptions(&pingMsg)
	if err != nil {
		log.Printf("Invalid ping options: %v", err)
		return
	}

	pingMsg.Address = formatAddress(pingMsg.Address)
	log.Printf("PING %s (%s): %d data bytes", pingMsg.Address, pingMsg.Address, opts.PacketSize)

	client := &http.Client{
		Timeout: time.Duration(opts.Timeout) * time.Second,
	}

	ticker := time.NewTicker(time.Duration(opts.Wait) * time.Second)
	defer ticker.Stop()

	sequence := 0

	if opts.Preload > 0 {
		for i := 0; i < opts.Preload; i++ {
			go func() {
				latency, err := measureLatency(client, pingMsg.Address)
				if err == nil {
					logPingResult(pingMsg.Address, -1, latency, true)
				}
			}()
		}
	}

	for range ticker.C {
		if opts.Count > 0 && sequence >= opts.Count {
			break
		}
		sequence++

		currentPacketSize := opts.PacketSize
		if opts.SweepMaxSize > 0 {
			currentPacketSize = opts.SweepMinSize +
				((sequence-1)*opts.SweepIncrSize)%
					(opts.SweepMaxSize-opts.SweepMinSize+1)
		}

		latency, err := measureLatency(client, pingMsg.Address)
		success := err == nil

		pong := createPongMessage(pingMsg.Address, sequence-1, latency, success)
		pong.Bytes = currentPacketSize

		if !opts.IsQuiet {
			if err := sendPongMessage(conn, pong); err != nil {
				log.Printf("Failed to send pong: %v", err)
				return
			}
		}

		if !opts.IsQuiet {
			logPingResult(pingMsg.Address, sequence-1, latency, success)
		}

		if !opts.IsFlood {
			if err := checkConnection(conn); err != nil {
				log.Printf("Connection check failed: %v", err)
				return
			}
		}

		if opts.IsFlood {
			time.Sleep(time.Millisecond)
		}
	}
}
