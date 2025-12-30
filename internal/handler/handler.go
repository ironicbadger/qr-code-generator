package handler

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ironicbadger/qr-code-generator/internal/qrcode"
	"github.com/ironicbadger/qr-code-generator/internal/storage"
)

type Handler struct {
	store     *storage.Store
	generator *qrcode.Generator
	templates *template.Template
}

func New(store *storage.Store, generator *qrcode.Generator, templates *template.Template) *Handler {
	return &Handler{
		store:     store,
		generator: generator,
		templates: templates,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("POST /generate", h.handleGenerate)
	mux.HandleFunc("GET /qr/{id}", h.handleGetQR)
	mux.HandleFunc("PUT /qr/{id}", h.handleUpdateLabel)
	mux.HandleFunc("DELETE /qr/{id}", h.handleDelete)
	mux.HandleFunc("GET /health", h.handleHealth)
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	codes, err := h.store.List(100, 0)
	if err != nil {
		log.Printf("Error listing QR codes: %v", err)
		http.Error(w, "Failed to load QR codes", http.StatusInternalServerError)
		return
	}

	data := struct {
		QRCodes []*storage.QRCode
	}{
		QRCodes: codes,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func (h *Handler) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	imageData, err := h.generator.Generate(content)
	if err != nil {
		log.Printf("Error generating QR code: %v", err)
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	_, err = h.store.Create(content, "", imageData)
	if err != nil {
		log.Printf("Error saving QR code: %v", err)
		http.Error(w, "Failed to save QR code", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) handleGetQR(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	qr, err := h.store.GetByID(id)
	if err != nil {
		log.Printf("Error getting QR code: %v", err)
		http.Error(w, "Failed to get QR code", http.StatusInternalServerError)
		return
	}
	if qr == nil {
		http.Error(w, "QR code not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"qr-"+idStr+".png\"")
	if _, err := w.Write(qr.ImageData); err != nil {
		log.Printf("Error writing QR image: %v", err)
	}
}

func (h *Handler) handleUpdateLabel(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Label string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateLabel(id, req.Label); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "QR code not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating label: %v", err)
		http.Error(w, "Failed to update label", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "QR code not found", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting QR code: %v", err)
		http.Error(w, "Failed to delete QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
