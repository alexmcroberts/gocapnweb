package gocapnweb

import (
	"bufio"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

// SetupRpcEndpoint sets up both WebSocket and HTTP POST endpoints for RPC using Echo.
func SetupRpcEndpoint(e *echo.Echo, path string, target RpcTarget) {
	session := NewRpcSession(target)

	// Setup WebSocket endpoint
	e.GET(path, func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return err
		}
		defer conn.Close()

		sessionData := NewSessionData(target)
		session.OnOpen(sessionData)
		defer session.OnClose(sessionData)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			response, err := session.HandleMessage(sessionData, string(message))
			if err != nil {
				log.Printf("Error processing WebSocket message: %v", err)
				continue
			}

			if response != "" {
				if err := conn.WriteMessage(websocket.TextMessage, []byte(response)); err != nil {
					log.Printf("Error writing WebSocket response: %v", err)
					break
				}
			}
		}
		return nil
	})

	// Setup HTTP POST endpoint for batch RPC
	e.POST(path, func(c echo.Context) error {
		// CORS headers are handled by Echo middleware
		c.Response().Header().Set("Content-Type", "text/plain")

		defer c.Request().Body.Close()
		scanner := bufio.NewScanner(c.Request().Body)

		// Create a session data for this HTTP batch request
		sessionData := NewSessionData(target)
		var responses []string

		// Process each line as a separate RPC message
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			response, err := session.HandleMessage(sessionData, line)
			if err != nil {
				log.Printf("Error processing HTTP message: %v", err)
				continue
			}

			if response != "" {
				responses = append(responses, response)
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading HTTP body: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error reading request body")
		}

		// Join responses with newlines
		responseBody := strings.Join(responses, "\n")
		return c.String(http.StatusOK, responseBody)
	})

	// OPTIONS endpoint is handled automatically by Echo CORS middleware
}

// SetupEchoServer creates and configures an Echo server with common middleware.
func SetupEchoServer() *echo.Echo {
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Hide Echo banner for cleaner output
	e.HideBanner = true

	return e
}
