package gocapnweb

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// SetupFileEndpoint sets up a static file server endpoint using Echo.
func SetupFileEndpoint(e *echo.Echo, urlPath string, fsRoot string) {
	// Clean the URL path to ensure it ends with a slash for proper matching
	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}

	// Create handler function
	fileHandler := func(c echo.Context) error {
		// Extract the file path from the URL
		requestPath := c.Request().URL.Path
		filePath := requestPath
		
		// Remove the base path prefix
		basePath := strings.TrimSuffix(urlPath, "/")
		if strings.HasPrefix(filePath, basePath) {
			filePath = filePath[len(basePath):]
		}

		// Remove leading slash from file path
		filePath = strings.TrimPrefix(filePath, "/")

		// Default to index.html for directory requests
		if filePath == "" || strings.HasSuffix(filePath, "/") {
			filePath = path.Join(filePath, "index.html")
		}

		// Construct full filesystem path
		fullPath := filepath.Join(fsRoot, filePath)

		// Security check: ensure the path is within fsRoot
		absRoot, err := filepath.Abs(fsRoot)
		if err != nil {
			log.Printf("Error getting absolute path for root: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			log.Printf("Error getting absolute path for file: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		// Ensure the resolved path is within the root directory
		if !strings.HasPrefix(absPath, absRoot) {
			log.Printf("Access denied for path outside root: %s", absPath)
			return echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}

		// Check if file exists and is a regular file
		fileInfo, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return echo.NewHTTPError(http.StatusNotFound, "File not found")
			} else {
				log.Printf("Error accessing file: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
			}
		}

		if !fileInfo.Mode().IsRegular() {
			return echo.NewHTTPError(http.StatusNotFound, "Not a file")
		}

		// Open and read the file
		file, err := os.Open(absPath)
		if err != nil {
			log.Printf("Error opening file: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read file")
		}
		defer file.Close()

		// Determine content type based on file extension
		contentType := getContentType(filepath.Ext(absPath))
		c.Response().Header().Set("Content-Type", contentType)

		// Set content length
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		// Copy file contents to response
		if _, err := io.Copy(c.Response(), file); err != nil {
			log.Printf("Error writing file to response: %v", err)
			return err
		}

		return nil
	}

	// Register the handler for the path pattern
	pathPattern := urlPath + "*"
	e.GET(pathPattern, fileHandler)
}

// getContentType returns the MIME type for a given file extension.
func getContentType(ext string) string {
	// First try the standard mime package
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	// Fallback to common types
	switch strings.ToLower(ext) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".mjs":
		return "text/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	default:
		return "application/octet-stream"
	}
}
