package handler

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"qr-code-generator/internal/qrcode"
	"qr-code-generator/internal/storage"
)

func setupTestHandler(t *testing.T) (*Handler, func()) {
	tmpDir, err := os.MkdirTemp("", "handler-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := storage.New(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create store: %v", err)
	}

	generator := qrcode.New()

	// Minimal template for testing
	tmpl := template.Must(template.New("index.html").Parse(`
		<!DOCTYPE html>
		<html>
		<body>
			{{range .QRCodes}}<div>{{.ID}}: {{.Content}}</div>{{end}}
		</body>
		</html>
	`))

	h := New(store, generator, tmpl)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return h, cleanup
}

func TestHandleHealth(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", resp["status"])
	}
}

func TestHandleIndex(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHandleGenerate(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	form := url.Values{}
	form.Set("content", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.handleGenerate(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status 303 (redirect), got %d", w.Code)
	}

	if w.Header().Get("Location") != "/" {
		t.Errorf("Expected redirect to /, got %s", w.Header().Get("Location"))
	}
}

func TestHandleGenerateEmpty(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	form := url.Values{}
	form.Set("content", "")

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.handleGenerate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleGetQR(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a QR code first
	qr, _ := h.store.Create("test", "", []byte{0x89, 0x50, 0x4E, 0x47})

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/qr/"+string(rune(qr.ID+'0')), nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.handleGetQR(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Expected Content-Type image/png, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHandleGetQRNotFound(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/qr/99999", nil)
	req.SetPathValue("id", "99999")
	w := httptest.NewRecorder()

	h.handleGetQR(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleUpdateLabel(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a QR code first
	qr, _ := h.store.Create("test", "", []byte{0x89, 0x50, 0x4E, 0x47})

	body := bytes.NewBufferString(`{"label":"Updated"}`)
	req := httptest.NewRequest(http.MethodPut, "/qr/1", body)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.handleUpdateLabel(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify update
	updated, _ := h.store.GetByID(qr.ID)
	if updated.Label != "Updated" {
		t.Errorf("Expected label 'Updated', got '%s'", updated.Label)
	}
}

func TestHandleDelete(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create a QR code first
	qr, _ := h.store.Create("test", "", []byte{0x89, 0x50, 0x4E, 0x47})

	req := httptest.NewRequest(http.MethodDelete, "/qr/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	h.handleDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify deletion
	deleted, _ := h.store.GetByID(qr.ID)
	if deleted != nil {
		t.Error("Expected QR code to be deleted")
	}
}
