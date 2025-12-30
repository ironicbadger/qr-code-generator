package qrcode

import (
	"bytes"
	"testing"
)

func TestGenerator(t *testing.T) {
	g := New()

	// Test Generate
	png, err := g.Generate("https://example.com")
	if err != nil {
		t.Fatalf("Failed to generate QR code: %v", err)
	}
	if len(png) == 0 {
		t.Error("Expected non-empty PNG data")
	}

	// Verify it's a valid PNG (starts with PNG magic bytes)
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47}
	if !bytes.HasPrefix(png, pngMagic) {
		t.Error("Generated data is not a valid PNG")
	}

	// Test Generate with empty content
	_, err = g.Generate("")
	if err == nil {
		t.Error("Expected error for empty content")
	}

	// Test GenerateWithSize
	png256, err := g.GenerateWithSize("test", 256)
	if err != nil {
		t.Fatalf("Failed to generate QR code with size: %v", err)
	}

	png512, err := g.GenerateWithSize("test", 512)
	if err != nil {
		t.Fatalf("Failed to generate QR code with size: %v", err)
	}

	// Larger size should generally produce larger file
	// (not always true due to compression, but for same content it typically is)
	if len(png512) <= len(png256) {
		t.Logf("Note: 512px PNG (%d bytes) not larger than 256px PNG (%d bytes)", len(png512), len(png256))
	}

	// Test GenerateWithSize with invalid size (should use default)
	png, err = g.GenerateWithSize("test", -1)
	if err != nil {
		t.Fatalf("Failed to generate QR code with invalid size: %v", err)
	}
	if len(png) == 0 {
		t.Error("Expected non-empty PNG data for invalid size")
	}

	// Test GenerateWithSize with empty content
	_, err = g.GenerateWithSize("", 256)
	if err == nil {
		t.Error("Expected error for empty content")
	}
}

func TestGeneratorLongContent(t *testing.T) {
	g := New()

	// Test with longer content (URL)
	longURL := "https://example.com/some/really/long/path/with/many/segments?param1=value1&param2=value2&param3=value3"
	png, err := g.Generate(longURL)
	if err != nil {
		t.Fatalf("Failed to generate QR code for long URL: %v", err)
	}
	if len(png) == 0 {
		t.Error("Expected non-empty PNG data")
	}
}
