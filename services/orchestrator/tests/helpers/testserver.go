package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// MockUserServiceServer creates a mock HTTP server for the User Service
type MockUserServiceServer struct {
	server         *httptest.Server
	mu             sync.RWMutex
	preferences    map[string]UserPreferencesResponse
	responseDelay  time.Duration
	errorRate      float64 // 0.0 to 1.0
	requestCount   int
	lastRequest    *http.Request
	customHandlers map[string]http.HandlerFunc
}

// UserPreferencesResponse represents the user service response format
type UserPreferencesResponse struct {
	EmailEnabled bool `json:"email_enabled"`
	PushEnabled  bool `json:"push_enabled"`
}

// NewMockUserServiceServer creates a new mock user service server
func NewMockUserServiceServer(t *testing.T) *MockUserServiceServer {
	mock := &MockUserServiceServer{
		preferences:    make(map[string]UserPreferencesResponse),
		customHandlers: make(map[string]http.HandlerFunc),
	}

	mux := http.NewServeMux()

	// Default preferences endpoint
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		mock.mu.Lock()
		mock.requestCount++
		mock.lastRequest = r
		mock.mu.Unlock()

		// Simulate delay
		if mock.responseDelay > 0 {
			time.Sleep(mock.responseDelay)
		}

		// Simulate errors
		if mock.errorRate > 0 && shouldError(mock.errorRate) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Service temporarily unavailable",
				},
			})
			return
		}

		// Extract user ID from path
		path := r.URL.Path
		userID := extractUserIDFromPath(path)

		mock.mu.RLock()
		prefs, exists := mock.preferences[userID]
		mock.mu.RUnlock()

		if !exists {
			// Default preferences
			prefs = UserPreferencesResponse{
				EmailEnabled: true,
				PushEnabled:  true,
			}
		}

		// Check for custom handler
		if handler, ok := mock.customHandlers[r.URL.Path]; ok {
			handler(w, r)
			return
		}

		// Return preferences
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prefs)
	})

	mock.server = httptest.NewServer(mux)
	return mock
}

// URL returns the base URL of the mock server
func (m *MockUserServiceServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockUserServiceServer) Close() {
	m.server.Close()
}

// SetPreferences sets preferences for a specific user
func (m *MockUserServiceServer) SetPreferences(userID string, prefs UserPreferencesResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.preferences[userID] = prefs
}

// SetResponseDelay sets a delay for all responses
func (m *MockUserServiceServer) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// SetErrorRate sets the error rate (0.0 to 1.0)
func (m *MockUserServiceServer) SetErrorRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorRate = rate
}

// GetRequestCount returns the number of requests received
func (m *MockUserServiceServer) GetRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestCount
}

// GetLastRequest returns the last received request
func (m *MockUserServiceServer) GetLastRequest() *http.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastRequest
}

// SetCustomHandler sets a custom handler for a specific path
func (m *MockUserServiceServer) SetCustomHandler(path string, handler http.HandlerFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customHandlers[path] = handler
}

// MockTemplateServiceServer creates a mock HTTP server for the Template Service
type MockTemplateServiceServer struct {
	server            *httptest.Server
	mu                sync.RWMutex
	templates         map[string]TemplateResponse
	renderedTemplates map[string]string
	responseDelay     time.Duration
	errorRate         float64
	requestCount      int
	lastRequest       *http.Request
	customHandlers    map[string]http.HandlerFunc
}

// TemplateResponse represents the template service response format
type TemplateResponse struct {
	ID        string   `json:"id"`
	Code      string   `json:"code"`
	Subject   string   `json:"subject"`
	Body      string   `json:"body"`
	Language  string   `json:"language"`
	Variables []string `json:"variables"`
}

// NewMockTemplateServiceServer creates a new mock template service server
func NewMockTemplateServiceServer(t *testing.T) *MockTemplateServiceServer {
	mock := &MockTemplateServiceServer{
		templates:         make(map[string]TemplateResponse),
		renderedTemplates: make(map[string]string),
		customHandlers:    make(map[string]http.HandlerFunc),
	}

	mux := http.NewServeMux()

	// Get template endpoint (GET /api/v1/templates/{template_id})
	// Render template endpoint (POST /api/v1/templates/{template_id}/render)
	mux.HandleFunc("/api/v1/templates/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mock.handleGetTemplate(w, r)
		} else if r.Method == http.MethodPost {
			// POST requests are for rendering templates
			mock.handleRenderTemplate(w, r)
		}
	})

	mock.server = httptest.NewServer(mux)
	return mock
}

func (m *MockTemplateServiceServer) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestCount++
	m.lastRequest = r
	m.mu.Unlock()

	if m.responseDelay > 0 {
		time.Sleep(m.responseDelay)
	}

	if m.errorRate > 0 && shouldError(m.errorRate) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "SERVICE_UNAVAILABLE",
				"message": "Service temporarily unavailable",
			},
		})
		return
	}

	templateID := extractTemplateIDFromPath(r.URL.Path)

	m.mu.RLock()
	template, exists := m.templates[templateID]
	m.mu.RUnlock()

	if !exists {
		// Default template
		template = TemplateResponse{
			ID:        templateID,
			Code:      templateID,
			Subject:   "Default Subject",
			Body:      "Default Body",
			Language:  "en",
			Variables: []string{"name"},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(template)
}

func (m *MockTemplateServiceServer) handleRenderTemplate(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.requestCount++
	m.lastRequest = r
	m.mu.Unlock()

	if m.responseDelay > 0 {
		time.Sleep(m.responseDelay)
	}

	// Extract template ID from URL path: /api/v1/templates/{template_id}/render
	templateID := extractTemplateIDFromPath(r.URL.Path)

	var req struct {
		Variables   map[string]interface{} `json:"variables"`
		Language    string                 `json:"language,omitempty"`
		Version     string                 `json:"version,omitempty"`
		PreviewMode bool                   `json:"preview_mode,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	m.mu.RLock()
	template, exists := m.templates[templateID]
	m.mu.RUnlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "TEMPLATE_NOT_FOUND",
				"message": fmt.Sprintf("Template %s not found", templateID),
			},
		})
		return
	}

	// Simple variable substitution
	response := map[string]interface{}{
		"rendered_subject": replaceVariable(template.Subject, req.Variables),
		"rendered_body":    replaceVariable(template.Body, req.Variables),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// URL returns the base URL of the mock server
func (m *MockTemplateServiceServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockTemplateServiceServer) Close() {
	m.server.Close()
}

// SetTemplate sets a template for testing
func (m *MockTemplateServiceServer) SetTemplate(templateID string, template TemplateResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[templateID] = template
}

// SetResponseDelay sets a delay for all responses
func (m *MockTemplateServiceServer) SetResponseDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responseDelay = delay
}

// SetErrorRate sets the error rate
func (m *MockTemplateServiceServer) SetErrorRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorRate = rate
}

// GetRequestCount returns the number of requests received
func (m *MockTemplateServiceServer) GetRequestCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requestCount
}

// Helper functions
func extractUserIDFromPath(path string) string {
	// Path format: /api/v1/users/{user_id}/preferences
	parts := splitPath(path)
	for i, part := range parts {
		if part == "users" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "default_user"
}

func extractTemplateIDFromPath(path string) string {
	// Path format: /api/v1/templates/{template_id} or /api/v1/templates/{template_id}/render
	parts := splitPath(path)
	for i, part := range parts {
		if part == "templates" && i+1 < len(parts) {
			templateID := parts[i+1]
			// Return the template ID (ignore "render" if it's the next part)
			return templateID
		}
	}
	return "default_template"
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, char := range path {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func replaceVariable(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

func replaceAll(s, old, new string) string {
	result := s
	for {
		replaced := replaceFirst(result, old, new)
		if replaced == result {
			break
		}
		result = replaced
	}
	return result
}

func replaceFirst(s, old, new string) string {
	if len(old) == 0 {
		return s
	}
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
	}
	return s
}

func shouldError(rate float64) bool {
	// Simple random error simulation
	// In real tests, you might want to use a proper random number generator
	return time.Now().UnixNano()%100 < int64(rate*100)
}
