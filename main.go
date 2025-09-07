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
	mux            *http.ServeMux
	templates      *template.Template
	config         *internal.Config
	sessionManager *internal.SessionManager
}

// NewServer creates a new server instance
func NewServer(config *internal.Config) (*Server, error) {
	// Parse templates
	templates, err := template.ParseFiles("templates/index.html", "templates/login.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		mux:            http.NewServeMux(),
		templates:      templates,
		config:         config,
		sessionManager: internal.NewSessionManager(),
	}
	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Serve static files (no auth required)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Auth routes (no auth required)
	s.mux.HandleFunc("/login", s.handleLogin)
	s.mux.HandleFunc("/logout", s.handleLogout)

	// Protected routes
	s.mux.HandleFunc("/", s.requireAuth(s.handleHome))
	s.mux.HandleFunc("/api/connections", s.requireAuth(s.handleConnectionsAPI))
	s.mux.HandleFunc("/api/connections/toggle", s.requireAuth(s.handleToggleAPI))
	s.mux.HandleFunc("/api/status", s.requireAuth(s.handleStatusAPI))
}

// handleHome serves the main HTML page
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", nil); err != nil {
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

// requireAuth middleware checks for valid authentication
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		_, valid := s.sessionManager.ValidateSession(cookie.Value)
		if !valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}

// handleLogin handles login form display and processing
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Show login form
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templateData := map[string]any{
			"Error": "",
		}
		if err := s.templates.ExecuteTemplate(w, "login.html", templateData); err != nil {
			log.Printf("Error rendering login template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	case http.MethodPost:
		password := r.FormValue("password")

		// Validate credentials
		if internal.ValidatePassword(password, s.config.PasswordHash) {
			// Create session
			sessionID, expires, err := s.sessionManager.CreateSession()
			if err != nil {
				log.Printf("Error creating session: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Set session cookie
			cookie := &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Expires:  expires,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			}
			http.SetCookie(w, cookie)

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			// Invalid credentials
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templateData := map[string]any{
				"Error": "Wrong password",
			}
			if err := s.templates.ExecuteTemplate(w, "login.html", templateData); err != nil {
				log.Printf("Error rendering login template: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLogout handles user logout
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie and delete session
	if cookie, err := r.Cookie("session_id"); err == nil {
		s.sessionManager.DeleteSession(cookie.Value)
	}

	// Clear session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.GetAddress()
	log.Printf("Starting on http://%s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}

func main() {
	// Load configuration
	config, err := internal.LoadConfig("config.yml")
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
