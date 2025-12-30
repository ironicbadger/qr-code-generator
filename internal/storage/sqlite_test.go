package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	// Create temp directory for test database
	tmpDir, err := os.MkdirTemp("", "qrcode-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	})

	dbPath := filepath.Join(tmpDir, "test.db")

	// Test New
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close store: %v", err)
		}
	})

	// Test Create
	imageData := []byte("test-image-data")
	qr, err := store.Create("https://example.com", "Test Label", imageData)
	if err != nil {
		t.Fatalf("Failed to create QR code: %v", err)
	}
	if qr.ID == 0 {
		t.Error("Expected non-zero ID")
	}
	if qr.Content != "https://example.com" {
		t.Errorf("Expected content 'https://example.com', got '%s'", qr.Content)
	}
	if qr.Label != "Test Label" {
		t.Errorf("Expected label 'Test Label', got '%s'", qr.Label)
	}

	// Test GetByID
	retrieved, err := store.GetByID(qr.ID)
	if err != nil {
		t.Fatalf("Failed to get QR code: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected non-nil QR code")
	}
	if string(retrieved.ImageData) != string(imageData) {
		t.Error("Image data mismatch")
	}

	// Test GetByID not found
	notFound, err := store.GetByID(99999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("Expected nil for non-existent ID")
	}

	// Test List
	// Add more QR codes
	if _, err := store.Create("content2", "", []byte("data2")); err != nil {
		t.Fatalf("Failed to create QR code: %v", err)
	}
	if _, err := store.Create("content3", "", []byte("data3")); err != nil {
		t.Fatalf("Failed to create QR code: %v", err)
	}

	codes, err := store.List(10, 0)
	if err != nil {
		t.Fatalf("Failed to list QR codes: %v", err)
	}
	if len(codes) != 3 {
		t.Errorf("Expected 3 QR codes, got %d", len(codes))
	}

	// Test List with limit
	codes, err = store.List(2, 0)
	if err != nil {
		t.Fatalf("Failed to list QR codes: %v", err)
	}
	if len(codes) != 2 {
		t.Errorf("Expected 2 QR codes, got %d", len(codes))
	}

	// Test UpdateLabel
	err = store.UpdateLabel(qr.ID, "Updated Label")
	if err != nil {
		t.Fatalf("Failed to update label: %v", err)
	}
	updated, err := store.GetByID(qr.ID)
	if err != nil {
		t.Fatalf("Failed to get QR code: %v", err)
	}
	if updated.Label != "Updated Label" {
		t.Errorf("Expected label 'Updated Label', got '%s'", updated.Label)
	}

	// Test UpdateLabel not found
	err = store.UpdateLabel(99999, "label")
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}

	// Test Delete
	err = store.Delete(qr.ID)
	if err != nil {
		t.Fatalf("Failed to delete QR code: %v", err)
	}
	deleted, err := store.GetByID(qr.ID)
	if err != nil {
		t.Fatalf("Failed to get QR code: %v", err)
	}
	if deleted != nil {
		t.Error("Expected nil after delete")
	}

	// Test Delete not found
	err = store.Delete(99999)
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}
}

func TestStoreCreateDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "qrcode-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	})

	// Create store with nested path that doesn't exist
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store with nested path: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("Failed to close store: %v", err)
		}
	})

	// Verify the directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}
