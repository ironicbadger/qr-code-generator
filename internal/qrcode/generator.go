package qrcode

import (
	"fmt"

	qr "github.com/skip2/go-qrcode"
)

const (
	DefaultSize    = 256
	ThumbnailSize  = 64
	RecoveryLevel  = qr.Medium
)

type Generator struct {
	size int
}

func New() *Generator {
	return &Generator{
		size: DefaultSize,
	}
}

func (g *Generator) Generate(content string) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	png, err := qr.Encode(content, RecoveryLevel, g.size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}

	return png, nil
}

func (g *Generator) GenerateWithSize(content string, size int) ([]byte, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	if size <= 0 {
		size = DefaultSize
	}

	png, err := qr.Encode(content, RecoveryLevel, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}

	return png, nil
}
