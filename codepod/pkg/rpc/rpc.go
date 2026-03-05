// Package rpc provides JSON-RPC 2.0 transport with Yamux multiplexing.
package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

const (
	// JSONRPCVersion is the supported JSON-RPC version.
	JSONRPCVersion = "2.0"

	// Error codes
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// JSON-RPC types
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc error: %d - %s", e.Code, e.Message)
}

// RPCServer represents a JSON-RPC server.
type RPCServer struct {
	mu       sync.RWMutex
	methods  map[string]reflect.Value
	handlers map[string]any
}

// NewRPCServer creates a new RPC server.
func NewRPCServer() *RPCServer {
	return &RPCServer{
		methods:  make(map[string]reflect.Value),
		handlers: make(map[string]any),
	}
}

// Register registers a method handler.
func (s *RPCServer) Register(name string, handler any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Convert handler to reflect.Value for faster calls
	s.methods[name] = reflect.ValueOf(handler)
	s.handlers[name] = handler
}

// HandleRequest processes a JSON-RPC request and returns the response.
func (s *RPCServer) HandleRequest(data []byte) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return s.errorResponse(nil, CodeParseError, "Parse error")
	}

	// Check JSON-RPC version
	if req.JSONRPC != JSONRPCVersion {
		return s.errorResponse(req.ID, CodeInvalidRequest, "Invalid JSON-RPC version")
	}

	// Find the method
	method, ok := s.methods[req.Method]
	if !ok {
		return s.errorResponse(req.ID, CodeMethodNotFound, "Method not found")
	}

	// Call the method
	result, err := s.callMethod(method, req.Params)
	if err != nil {
		return s.errorResponse(req.ID, CodeInvalidParams, err.Error())
	}

	// Return success response
	resultBytes, _ := json.Marshal(result)
	return s.successResponse(req.ID, resultBytes)
}

// callMethod invokes the handler with the given params.
func (s *RPCServer) callMethod(method reflect.Value, params json.RawMessage) (any, error) {
	// Get handler type
	handlerType := method.Type()

	// Check if it's a function
	if handlerType.Kind() != reflect.Func {
		return nil, errors.New("handler is not a function")
	}

	// Get the function to call
	fn := method.Interface()

	// Simple case: function with (context.Context, params) returns (result, error)
	switch fn := fn.(type) {
	case func(context.Context, any) (any, error):
		var paramsVal any
		if len(params) > 0 {
			if err := json.Unmarshal(params, &paramsVal); err != nil {
				return nil, err
			}
		}
		return fn(context.Background(), paramsVal)

	case func(any) (any, error):
		var paramsVal any
		if len(params) > 0 {
			if err := json.Unmarshal(params, &paramsVal); err != nil {
				return nil, err
			}
		}
		return fn(paramsVal)

	case func() (any, error):
		return fn()

	default:
		return nil, errors.New("unsupported function signature")
	}
}

func (s *RPCServer) successResponse(id interface{}, result []byte) []byte {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		Result:  result,
		ID:      id,
	}
	data, _ := json.Marshal(resp)
	return data
}

func (s *RPCServer) errorResponse(id interface{}, code int, message string) []byte {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		Error: &Error{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	data, _ := json.Marshal(resp)
	return data
}

// RPCClient represents a JSON-RPC client.
type RPCClient struct {
	mu         sync.Mutex
	send       func(data []byte) error
	pending    map[interface{}]chan []byte
	notifications chan Notification
}

// Notification represents a notification from the server.
type Notification struct {
	Method string
	Params any
}

// NewRPCClient creates a new RPC client.
func NewRPCClient(sendFunc func(data []byte) error) *RPCClient {
	return &RPCClient{
		send:         sendFunc,
		pending:      make(map[interface{}]chan []byte),
		notifications: make(chan Notification, 10),
	}
}

// Call sends a request and waits for response.
func (c *RPCClient) Call(ctx context.Context, method string, params any) (any, error) {
	c.mu.Lock()

	// Generate ID
	id := generateID()
	resultCh := make(chan []byte, 1)
	c.pending[id] = resultCh

	// Build request
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		ID:      id,
	}
	if params != nil {
		paramsBytes, _ := json.Marshal(params)
		req.Params = paramsBytes
	}

	data, _ := json.Marshal(req)

	c.mu.Unlock()

	// Send request
	if err := c.send(data); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	// Wait for response
	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case result := <-resultCh:
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()

		// Parse response
		var resp Response
		if err := json.Unmarshal(result, &resp); err != nil {
			return nil, err
		}

		if resp.Error != nil {
			return nil, resp.Error
		}

		if resp.Result == nil {
			return nil, nil
		}

		var resultVal any
		if err := json.Unmarshal(resp.Result, &resultVal); err != nil {
			return nil, err
		}
		return resultVal, nil
	}
}

// Notify sends a notification (no response expected).
func (c *RPCClient) Notify(method string, params any) error {
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
	}
	if params != nil {
		paramsBytes, _ := json.Marshal(params)
		req.Params = paramsBytes
	}

	data, _ := json.Marshal(req)
	return c.send(data)
}

// HandleResponse processes a response from the server.
func (c *RPCClient) HandleResponse(data []byte) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return
	}

	// Check if it's a notification (no ID)
	if resp.ID == nil {
		// Handle notification - need to parse from original data
		// Re-parse as request to get method
		var notif Request
		if err := json.Unmarshal(data, &notif); err == nil {
			var params any
			if notif.Params != nil {
				json.Unmarshal(notif.Params, &params)
			}
			select {
			case c.notifications <- Notification{Method: notif.Method, Params: params}:
			default:
			}
		}
		return
	}

	// Find pending request
	c.mu.Lock()
	ch, ok := c.pending[resp.ID]
	c.mu.Unlock()

	if ok {
		ch <- data
	}
}

// Close cleans up the client.
func (c *RPCClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	return nil
}

// ID generator
var idCounter uint64
var idMu sync.Mutex

func generateID() interface{} {
	idMu.Lock()
	defer idMu.Unlock()
	idCounter++
	return idCounter
}
