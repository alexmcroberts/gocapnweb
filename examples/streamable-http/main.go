package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gocapnweb"
	"github.com/labstack/echo/v4"
)

// Simulated "tokens" for the typewriter effect. In a real app these would come from an LLM.
var samplePhrases = []string{
	"The", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog.",
	"Streamable", "HTTP", "sends", "each", "chunk", "as", "soon", "as", "it's", "ready.",
	"No", "need", "to", "wait", "for", "the", "entire", "response.",
	"Perfect", "for", "AI", "typewriters,", "log", "tails,", "and", "live", "feeds.",
}

func main() {
	staticPath := "./static"
	if len(os.Args) >= 2 {
		staticPath = os.Args[1]
	}
	port := ":8000"

	e := gocapnweb.SetupEchoServer()

	// Streaming text generation endpoint: reads prompt from body, streams response line-by-line.
	e.POST("/api/generate-stream", func(c echo.Context) error {
		// Read prompt (first line or full body)
		var prompt string
		scanner := bufio.NewScanner(c.Request().Body)
		if scanner.Scan() {
			prompt = strings.TrimSpace(scanner.Text())
		}
		if prompt == "" {
			prompt = "Tell me a story"
		}
		_ = prompt // could be used to customize generation

		// Stream response with chunked transfer encoding
		c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
		c.Response().WriteHeader(http.StatusOK)

		flusher, ok := c.Response().Writer.(http.Flusher)
		if !ok {
			return c.String(http.StatusInternalServerError, "streaming not supported")
		}

		// Simulate token-by-token generation: write each token and flush
		for i, token := range samplePhrases {
			line := token
			if i < len(samplePhrases)-1 && !strings.HasSuffix(token, ",") && !strings.HasSuffix(token, ".") {
				line += " "
			}
			if _, err := c.Response().Write([]byte(line)); err != nil {
				return err
			}
			flusher.Flush()
			time.Sleep(80 * time.Millisecond)
		}

		return nil
	})

	gocapnweb.SetupFileEndpoint(e, "/static", staticPath)

	log.Printf("Streamable HTTP example server starting on %s", port)
	log.Printf("Static files: %s", staticPath)
	log.Printf("POST /api/generate-stream — streaming text generation")
	log.Printf("Demo: start Svelte dev server in static/ then open http://localhost:3000")

	if err := e.Start(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
