// Package gocapnweb provides a Go implementation of the Cap'n Web RPC protocol server.
// This library allows creating server implementations for the Cap'n Web RPC protocol
// with support for WebSocket and HTTP batch endpoints.
package gocapnweb

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// RpcTarget defines the interface that server implementations must satisfy.
// It provides method dispatch functionality for incoming RPC calls.
type RpcTarget interface {
	// Dispatch handles method calls and returns the result as JSON.
	// It should return an error if the method is not found or execution fails.
	Dispatch(method string, args json.RawMessage) (interface{}, error)
}

// BaseRpcTarget provides a convenient base implementation of RpcTarget
// with method registration capabilities.
type BaseRpcTarget struct {
	methods map[string]func(json.RawMessage) (interface{}, error)
	mu      sync.RWMutex
}

// NewBaseRpcTarget creates a new BaseRpcTarget instance.
func NewBaseRpcTarget() *BaseRpcTarget {
	return &BaseRpcTarget{
		methods: make(map[string]func(json.RawMessage) (interface{}, error)),
	}
}

// Method registers a method handler with the given name.
func (t *BaseRpcTarget) Method(name string, handler func(json.RawMessage) (interface{}, error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.methods[name] = handler
}

// Dispatch implements the RpcTarget interface.
func (t *BaseRpcTarget) Dispatch(method string, args json.RawMessage) (interface{}, error) {
	t.mu.RLock()
	handler, exists := t.methods[method]
	t.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	return handler(args)
}

// SessionData holds the state for each RPC session (WebSocket connection or HTTP batch).
type SessionData struct {
	PendingResults    map[int]interface{} `json:"pendingResults"`
	PendingOperations map[int]Operation   `json:"pendingOperations"`
	NextExportID      int                 `json:"nextExportId"`
	Target            RpcTarget           `json:"-"`
	mu                sync.RWMutex
}

// Operation represents a pending RPC operation.
type Operation struct {
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
}

// NewSessionData creates a new SessionData instance.
func NewSessionData(target RpcTarget) *SessionData {
	return &SessionData{
		PendingResults:    make(map[int]interface{}),
		PendingOperations: make(map[int]Operation),
		NextExportID:      1,
		Target:            target,
	}
}

// RpcSession handles the Cap'n Web RPC protocol for connections.
type RpcSession struct {
	target RpcTarget
}

// NewRpcSession creates a new RpcSession with the given target.
func NewRpcSession(target RpcTarget) *RpcSession {
	return &RpcSession{target: target}
}

// HandleMessage processes an incoming RPC message and returns the response.
// Returns an empty string if no response should be sent.
func (s *RpcSession) HandleMessage(sessionData *SessionData, message string) (string, error) {
	var msg []interface{}
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		return "", fmt.Errorf("invalid message format: %w", err)
	}

	if len(msg) == 0 {
		return "", fmt.Errorf("empty message")
	}

	messageType, ok := msg[0].(string)
	if !ok {
		return "", fmt.Errorf("invalid message type")
	}

	switch messageType {
	case "push":
		if len(msg) >= 2 {
			s.handlePush(sessionData, msg[1])
		}
		return "", nil // No response for push

	case "pull":
		if len(msg) >= 2 {
			if exportIDFloat, ok := msg[1].(float64); ok {
				exportID := int(exportIDFloat)
				response, err := s.handlePull(sessionData, exportID)
				if err != nil {
					return "", err
				}
				responseBytes, err := json.Marshal(response)
				if err != nil {
					return "", err
				}
				return string(responseBytes), nil
			}
		}

	case "release":
		if len(msg) >= 3 {
			if exportIDFloat, ok := msg[1].(float64); ok {
				if refcountFloat, ok := msg[2].(float64); ok {
					s.handleRelease(sessionData, int(exportIDFloat), int(refcountFloat))
				}
			}
		}
		return "", nil // No response for release

	case "abort":
		if len(msg) >= 2 {
			s.handleAbort(sessionData, msg[1])
		}
		return "", nil // No response for abort
	}

	return "", nil
}

// OnOpen initializes a new session.
func (s *RpcSession) OnOpen(sessionData *SessionData) {
	log.Println("WebSocket connection opened")
	sessionData.mu.Lock()
	defer sessionData.mu.Unlock()
	sessionData.NextExportID = 1
	sessionData.PendingResults = make(map[int]interface{})
	sessionData.PendingOperations = make(map[int]Operation)
}

// OnClose cleans up a session.
func (s *RpcSession) OnClose(sessionData *SessionData) {
	log.Println("WebSocket connection closed")
}

func (s *RpcSession) handlePush(sessionData *SessionData, pushData interface{}) {
	pushArray, ok := pushData.([]interface{})
	if !ok || len(pushArray) == 0 {
		return
	}

	sessionData.mu.Lock()
	defer sessionData.mu.Unlock()

	// Create a new export on the server side
	exportID := sessionData.NextExportID
	sessionData.NextExportID++

	if len(pushArray) >= 3 && pushArray[0] == "pipeline" {
		if importIDFloat, ok := pushArray[1].(float64); ok {
			_ = int(importIDFloat) // importID for future use

			if methodArray, ok := pushArray[2].([]interface{}); ok && len(methodArray) > 0 {
				if method, ok := methodArray[0].(string); ok {
					var args json.RawMessage
					if len(pushArray) >= 4 {
						argsBytes, _ := json.Marshal(pushArray[3])
						args = argsBytes
					} else {
						args = json.RawMessage("[]")
					}

					// Store the operation for lazy evaluation when pulled
					sessionData.PendingOperations[exportID] = Operation{
						Method: method,
						Args:   args,
					}
				}
			}
		}
	}
}

func (s *RpcSession) resolvePipelineReferences(sessionData *SessionData, value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		// Check if this is a pipeline reference: ["pipeline", exportId, ["path", ...]]
		if len(v) >= 2 {
			if pipelineStr, ok := v[0].(string); ok && pipelineStr == "pipeline" {
				if refExportIDFloat, ok := v[1].(float64); ok {
					refExportID := int(refExportIDFloat)

					sessionData.mu.RLock()
					// Check if result is already computed
					if result, exists := sessionData.PendingResults[refExportID]; exists {
						sessionData.mu.RUnlock()

						// If there's a path, traverse it
						if len(v) >= 3 {
							if pathArray, ok := v[2].([]interface{}); ok {
								return s.traversePath(result, pathArray)
							}
						}
						return result, nil
					}

					// Check if we need to execute a pending operation
					if operation, exists := sessionData.PendingOperations[refExportID]; exists {
						sessionData.mu.RUnlock()

						// Recursively resolve arguments
						var args interface{}
						if err := json.Unmarshal(operation.Args, &args); err != nil {
							return nil, err
						}
						resolvedArgs, err := s.resolvePipelineReferences(sessionData, args)
						if err != nil {
							return nil, err
						}

						resolvedArgsBytes, err := json.Marshal(resolvedArgs)
						if err != nil {
							return nil, err
						}

						// Execute the operation
						result, err := sessionData.Target.Dispatch(operation.Method, resolvedArgsBytes)
						if err != nil {
							return nil, err
						}

						// Cache the result
						sessionData.mu.Lock()
						sessionData.PendingResults[refExportID] = result
						delete(sessionData.PendingOperations, refExportID)
						sessionData.mu.Unlock()

						// If there's a path, traverse it
						if len(v) >= 3 {
							if pathArray, ok := v[2].([]interface{}); ok {
								return s.traversePath(result, pathArray)
							}
						}
						return result, nil
					}
					sessionData.mu.RUnlock()

					return nil, fmt.Errorf("pipeline reference to non-existent export: %d", refExportID)
				}
			}
		}

		// Not a pipeline reference, recursively resolve elements
		resolved := make([]interface{}, len(v))
		for i, elem := range v {
			resolvedElem, err := s.resolvePipelineReferences(sessionData, elem)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedElem
		}
		return resolved, nil

	case map[string]interface{}:
		// Recursively resolve object values
		resolved := make(map[string]interface{})
		for key, val := range v {
			resolvedVal, err := s.resolvePipelineReferences(sessionData, val)
			if err != nil {
				return nil, err
			}
			resolved[key] = resolvedVal
		}
		return resolved, nil

	default:
		// Primitive value, return as-is
		return value, nil
	}
}

func (s *RpcSession) traversePath(result interface{}, path []interface{}) (interface{}, error) {
	current := result
	for _, key := range path {
		switch k := key.(type) {
		case string:
			if obj, ok := current.(map[string]interface{}); ok {
				current = obj[k]
			} else {
				return nil, fmt.Errorf("cannot traverse string key on non-object")
			}
		case float64:
			if arr, ok := current.([]interface{}); ok {
				idx := int(k)
				if idx < 0 || idx >= len(arr) {
					return nil, fmt.Errorf("array index out of bounds")
				}
				current = arr[idx]
			} else {
				return nil, fmt.Errorf("cannot traverse numeric key on non-array")
			}
		default:
			return nil, fmt.Errorf("invalid path key type")
		}
	}
	return current, nil
}

func (s *RpcSession) handlePull(sessionData *SessionData, exportID int) ([]interface{}, error) {
	sessionData.mu.RLock()
	// Check if we already have a cached result
	if result, exists := sessionData.PendingResults[exportID]; exists {
		sessionData.mu.RUnlock()

		// Clean up
		sessionData.mu.Lock()
		delete(sessionData.PendingResults, exportID)
		sessionData.mu.Unlock()

		// Check if the stored result is an error
		if errArray, ok := result.([]interface{}); ok && len(errArray) >= 2 {
			if errType, ok := errArray[0].(string); ok && errType == "error" {
				// Send as reject
				return []interface{}{"reject", exportID, result}, nil
			}
		}

		// Send as resolve
		// Arrays need to be wrapped in another array to escape them per Cap'n Web protocol
		if _, ok := result.([]interface{}); ok {
			return []interface{}{"resolve", exportID, []interface{}{result}}, nil
		}
		return []interface{}{"resolve", exportID, result}, nil
	}

	// Check if we have a pending operation to execute
	if operation, exists := sessionData.PendingOperations[exportID]; exists {
		sessionData.mu.RUnlock()

		// Resolve any pipeline references in the arguments
		var args interface{}
		if err := json.Unmarshal(operation.Args, &args); err != nil {
			return s.createErrorResponse(exportID, "ArgumentError", err.Error()), nil
		}

		resolvedArgs, err := s.resolvePipelineReferences(sessionData, args)
		if err != nil {
			return s.createErrorResponse(exportID, "PipelineError", err.Error()), nil
		}

		resolvedArgsBytes, err := json.Marshal(resolvedArgs)
		if err != nil {
			return s.createErrorResponse(exportID, "SerializationError", err.Error()), nil
		}

		// Dispatch the method call to the target
		result, err := sessionData.Target.Dispatch(operation.Method, resolvedArgsBytes)

		// Clean up the operation
		sessionData.mu.Lock()
		delete(sessionData.PendingOperations, exportID)
		sessionData.mu.Unlock()

		if err != nil {
			return s.createErrorResponse(exportID, "MethodError", err.Error()), nil
		}

		// Store the result for future reference
		sessionData.mu.Lock()
		sessionData.PendingResults[exportID] = result
		sessionData.mu.Unlock()

		// Send as resolve
		if _, ok := result.([]interface{}); ok {
			return []interface{}{"resolve", exportID, []interface{}{result}}, nil
		}
		return []interface{}{"resolve", exportID, result}, nil
	}
	sessionData.mu.RUnlock()

	// Export ID not found - send an error
	return []interface{}{"reject", exportID, []interface{}{
		"error", "ExportNotFound", "Export ID not found",
	}}, nil
}

func (s *RpcSession) createErrorResponse(exportID int, errorType, message string) []interface{} {
	return []interface{}{"reject", exportID, []interface{}{
		"error", errorType, message,
	}}
}

func (s *RpcSession) handleRelease(sessionData *SessionData, exportID, refcount int) {
	log.Printf("Released export %d with refcount %d", exportID, refcount)
}

func (s *RpcSession) handleAbort(sessionData *SessionData, errorData interface{}) {
	errorBytes, _ := json.Marshal(errorData)
	log.Printf("Abort received: %s", string(errorBytes))
}
