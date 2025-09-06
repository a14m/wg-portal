package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"wg-portal/internal"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Server encapsulates our HTTP server
type Server struct {
	mux       *http.ServeMux
	templates *template.Template
	config    *internal.Config
}

// NewServer creates a new server instance
func NewServer(config *internal.Config) (*Server, error) {
	// Parse templates
	templates, err := template.ParseFiles("templates/index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		mux:       http.NewServeMux(),
		templates: templates,
		config:    config,
	}
	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Serve static files
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Main page
	s.mux.HandleFunc("/", s.handleHome)

	// API endpoints
	s.mux.HandleFunc("/api/connections", s.handleConnectionsAPI)
	s.mux.HandleFunc("/api/connections/toggle", s.handleToggleAPI)
	s.mux.HandleFunc("/api/status", s.handleStatusAPI)
}

// handleHome serves the main HTML page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templateData := map[string]any{
		"Title": s.config.Server.Title,
	}
	if err := s.templates.ExecuteTemplate(w, "index.html", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// handleConnectionsAPI returns connection data as JSON
func (s *Server) handleConnectionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connections, err := internal.GetConnections()
	if err != nil {
		log.Printf("Error getting connections: %v", err)
		s.sendErrorResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	s.sendSuccessResponse(w, connections)
}

// handleToggleAPI handles connection toggle requests
func (s *Server) handleToggleAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		s.sendErrorResponse(w, "Connection name is required", http.StatusBadRequest)
		return
	}

	output, err := internal.ToggleConnection(req.Name)
	if err != nil {
		log.Printf("Error toggling connection %s: %v (output: %s)", req.Name, err, string(output))
		s.sendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message": fmt.Sprintf("Connection %s toggled successfully", req.Name),
		"output":  string(output),
	}

	s.sendSuccessResponse(w, response)
}

// handleStatusAPI returns WireGuard status information
func (s *Server) handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status, err := internal.GetStatus()
	if err != nil {
		log.Printf("Error getting status: %v", err)
		s.sendErrorResponse(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"status": status,
	}

	s.sendSuccessResponse(w, response)
}

// sendSuccessResponse sends a JSON success response
func (*Server) sendSuccessResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// sendErrorResponse sends a JSON error response
func (*Server) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.GetAddress()
	log.Printf("%s starting on http://%s\n", s.config.Server.Title, addr)
	return http.ListenAndServe(addr, s.mux)
}

func main() {
	// Load configuration
	config, err := internal.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
