// Package main — HTTP server with embedded frontend and JSON-RPC API.
// The single binary serves the React UI, handles API calls, and manages PLM projects.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/Homiakus/go-plm/frontend"
	wailsapi "github.com/Homiakus/go-plm/internal/api/wails"
	"github.com/Homiakus/go-plm/internal/app/service"
)

// JSONRPCRequest is a standard JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is a standard JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server handles HTTP and serves the PLM application.
type Server struct {
	app        *service.App
	api        *wailsapi.API
	httpServer *http.Server
	root       string
	apiToken   string
}

// NewServer creates a server for the given project root.
func NewServer(root string) (*Server, error) {
	absRoot, err := resolveRoot(root)
	if err != nil {
		return nil, err
	}

	app, err := service.Open(absRoot)
	if err != nil {
		return nil, fmt.Errorf("open project at %s: %w", absRoot, err)
	}

	api := wailsapi.NewAPI(app)

	token, err := newAPIToken()
	if err != nil {
		app.Close()
		return nil, fmt.Errorf("create api token: %w", err)
	}

	srv := &Server{
		app:      app,
		api:      api,
		root:     absRoot,
		apiToken: token,
	}

	return srv, nil
}

// Start begins listening and opens the browser.
func (s *Server) Start(port int) error {
	mux := http.NewServeMux()

	// JSON-RPC endpoint
	mux.HandleFunc("/api", s.localAPIMiddleware(s.handleJSONRPC))

	// Static frontend (embedded)
	frontendFS, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		return fmt.Errorf("frontend not built — run 'cd frontend && npm run build' first: %w", err)
	}
	fileServer := http.FileServer(http.FS(frontendFS))
	mux.Handle("/", s.frontendHandler(fileServer))

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	url := fmt.Sprintf("http://%s", addr)
	fmt.Printf("\n")
	fmt.Printf("╔══════════════════════════════════════════════╗\n")
	fmt.Printf("║   go-plm v2.0 — Local PLM/PDM System        ║\n")
	fmt.Printf("╠══════════════════════════════════════════════╣\n")
	fmt.Printf("║  Project: %-34s ║\n", truncate(s.app.Config.Project.Title, 34))
	fmt.Printf("║  Root:    %-34s ║\n", truncate(s.root, 34))
	fmt.Printf("║  URL:     %-34s ║\n", truncate(url, 34))
	fmt.Printf("╚══════════════════════════════════════════════╝\n")
	fmt.Printf("\n  Press Ctrl+C to stop.\n\n")

	// Auto-open browser
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	// Serve
	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s.httpServer != nil {
		s.httpServer.Shutdown(ctx)
	}
	if s.api != nil {
		s.api.Close()
	}
}

// handleJSONRPC dispatches JSON-RPC 2.0 calls to API methods.
func (s *Server) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      0,
			Error:   &RPCError{Code: -32700, Message: "Parse error: " + err.Error()},
		})
		return
	}

	result, err := s.callMethod(req.Method, req.Params)

	if err != nil {
		writeJSON(w, http.StatusOK, JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32000, Message: err.Error()},
		})
		return
	}

	writeJSON(w, http.StatusOK, JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	})
}

// callMethod uses reflection to dispatch JSON-RPC method to API.
func (s *Server) callMethod(method string, params json.RawMessage) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("method %q failed: %v", method, r)
		}
	}()

	if !allowedRPCMethods[method] {
		return nil, fmt.Errorf("method %q not allowed", method)
	}

	apiVal := reflect.ValueOf(s.api)
	methodVal := apiVal.MethodByName(method)
	if !methodVal.IsValid() {
		return nil, fmt.Errorf("method %q not found", method)
	}

	methodType := methodVal.Type()

	// Build argument list from params
	args := make([]reflect.Value, methodType.NumIn())

	if methodType.NumIn() > 0 && len(params) > 0 {
		// Try to unmarshal params as array or single object
		firstParam := methodType.In(0)

		// If first param is a struct, unmarshal params as object
		if firstParam.Kind() == reflect.Struct || (firstParam.Kind() == reflect.Ptr && firstParam.Elem().Kind() == reflect.Struct) {
			if firstParam.Kind() == reflect.Ptr {
				val := reflect.New(firstParam.Elem())
				if err := json.Unmarshal(params, val.Interface()); err != nil {
					return nil, fmt.Errorf("invalid params for %s: %w", method, err)
				}
				args[0] = val
			} else {
				val := reflect.New(firstParam)
				if err := json.Unmarshal(params, val.Interface()); err != nil {
					return nil, fmt.Errorf("invalid params for %s: %w", method, err)
				}
				args[0] = val.Elem()
			}
		} else if firstParam.Kind() == reflect.String && len(params) > 0 {
			// Unwrap string from JSON array or string
			var arr []json.RawMessage
			if json.Unmarshal(params, &arr) == nil && len(arr) > 0 {
				for i, raw := range arr {
					if i >= methodType.NumIn() {
						break
					}
					pt := methodType.In(i)
					val := reflect.New(pt)
					if err := json.Unmarshal(raw, val.Interface()); err != nil {
						return nil, fmt.Errorf("param %d: %w", i, err)
					}
					args[i] = val.Elem()
				}
			} else {
				// Try as simple value
				val := reflect.New(firstParam)
				if err := json.Unmarshal(params, val.Interface()); err != nil {
					return nil, fmt.Errorf("invalid param for %s: %w", method, err)
				}
				args[0] = val.Elem()
			}
		} else {
			// Try array unwrap
			var arr []json.RawMessage
			if json.Unmarshal(params, &arr) == nil {
				for i, raw := range arr {
					if i >= methodType.NumIn() {
						break
					}
					pt := methodType.In(i)
					val := reflect.New(pt)
					if err := json.Unmarshal(raw, val.Interface()); err != nil {
						return nil, fmt.Errorf("param %d: %w", i, err)
					}
					args[i] = val.Elem()
				}
			} else {
				// Single value
				val := reflect.New(firstParam)
				if err := json.Unmarshal(params, val.Interface()); err != nil {
					return nil, fmt.Errorf("invalid param for %s: %w", method, err)
				}
				args[0] = val.Elem()
			}
		}
	}
	for i := range args {
		if !args[i].IsValid() {
			args[i] = reflect.Zero(methodType.In(i))
		}
	}

	// Call the method
	results := methodVal.Call(args)

	// Handle returns: (value, error) or (value) or ()
	if len(results) == 2 {
		// (value, error)
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}
		return results[0].Interface(), nil
	} else if len(results) == 1 {
		if methodType.Out(0).Kind() == reflect.Interface {
			// Might be error
			if err, ok := results[0].Interface().(error); ok {
				return nil, err
			}
		}
		return results[0].Interface(), nil
	}

	return nil, nil
}

var allowedRPCMethods = map[string]bool{
	"OpenProject":           true,
	"CreateProject":         true,
	"CreateObject":          true,
	"GetObject":             true,
	"ListObjects":           true,
	"SearchObjects":         true,
	"DeleteObject":          true,
	"RunTransition":         true,
	"AvailableTransitions":  true,
	"GetBOM":                true,
	"ValidateBOM":           true,
	"DetectBOMCycles":       true,
	"CheckReleaseReadiness": true,
	"BuildReleaseScope":     true,
	"GetTree":               true,
	"CreateCheckpoint":      true,
	"GitStatus":             true,
	"RebuildIndex":          true,
	"Stats":                 true,
	"ProjectInfo":           true,
	"GetObjectDocument":     true,
	"UpdateObjectDocument":  true,
	"ValidateObject":        true,
	"GetObjectHistory":      true,
	"GetNamingInfo":         true,
	"AttachArtifact":        true,
	"UpdateObject":          true,
	"DuplicateObject":       true,
	"CreateReleasePackage":  true,
	"GetNextSequence":       true,
	"WhereUsed":             true,
	"ExportBOM":             true,
	"BumpRevision":          true,
	"AddRelation":           true,
	"RecoverTransactions":   true,
	"RollbackTransaction":   true,
}

const apiTokenCookie = "plm_api_token"

// frontendHandler serves the UI and sets the per-process API token cookie.
func (s *Server) frontendHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.setAPITokenCookie(w)
		next.ServeHTTP(w, r)
	})
}

// localAPIMiddleware rejects non-local origins and requires the UI-issued API token.
func (s *Server) localAPIMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && !isAllowedOrigin(origin, r.Host) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if !s.hasValidAPIToken(r) {
			http.Error(w, "api token required", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (s *Server) setAPITokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     apiTokenCookie,
		Value:    s.apiToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) hasValidAPIToken(r *http.Request) bool {
	token := r.Header.Get("X-PLM-Token")
	if token == "" {
		if cookie, err := r.Cookie(apiTokenCookie); err == nil {
			token = cookie.Value
		}
	}
	return constantTimeEqual(token, s.apiToken)
}

func constantTimeEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func newAPIToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func isAllowedOrigin(origin, host string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, host)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func resolveRoot(path string) (string, error) {
	if path == "" || path == "." {
		return os.Getwd()
	}
	return path, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

// Ensure embed/frontend is used
var _ = frontend.Assets
