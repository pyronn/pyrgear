package comands

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessImageExif(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "exif_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		assert.NoError(t, err)
	}()

	// Test with non-existent file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.jpg")
	err = processImageExif(nonExistentFile, "text")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	// Test with unsupported file format
	txtFile := filepath.Join(tempDir, "test.txt")
	f, err := os.Create(txtFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = f.Close()
	assert.NoError(t, err)

	err = processImageExif(txtFile, "text")
	if err == nil {
		t.Error("Expected error for unsupported file format, got nil")
	}
}

func TestProcessDirectoryExif(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "exif_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		assert.NoError(t, err)
	}()

	// Test with non-existent directory
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	err = processDirectoryExif(nonExistentDir, "text", false)
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}

	// Test with valid directory (empty)
	err = processDirectoryExif(tempDir, "text", false)
	if err != nil {
		t.Errorf("Unexpected error for empty directory: %v", err)
	}

	// Create a test file (not an image)
	txtFile := filepath.Join(tempDir, "test.txt")
	f, err := os.Create(txtFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = f.Close()
	assert.NoError(t, err)

	// Test with directory containing non-image files
	err = processDirectoryExif(tempDir, "text", false)
	if err != nil {
		t.Errorf("Unexpected error for directory with non-image files: %v", err)
	}
}

func TestSupportedImageFormats(t *testing.T) {
	supportedExts := []string{".jpg", ".jpeg", ".tiff", ".tif"}
	unsupportedExts := []string{".png", ".gif", ".bmp", ".webp"}

	tempDir, err := os.MkdirTemp("", "format_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		assert.NoError(t, err)
	}()

	// Test supported formats (will fail due to invalid EXIF data, but format should be accepted)
	for _, ext := range supportedExts {
		testFile := filepath.Join(tempDir, "test"+ext)
		f, err := os.Create(testFile)
		assert.NoError(t, err)
		_, err = f.WriteString("fake image data")
		assert.NoError(t, err)
		assert.NoError(t, f.Close())

		err = processImageExif(testFile, "text")
		// We expect an EXIF decode error, but not a format error
		if err != nil && !containsString(err.Error(), "failed to decode EXIF data") {
			t.Errorf("Unexpected error type for supported format %s: %v", ext, err)
		}
	}

	// Test unsupported formats
	for _, ext := range unsupportedExts {
		testFile := filepath.Join(tempDir, "test"+ext)
		f, err := os.Create(testFile)
		assert.NoError(t, err)
		err = f.Close()
		assert.NoError(t, err)

		err = processImageExif(testFile, "text")
		if err == nil || !containsString(err.Error(), "unsupported image format") {
			t.Errorf("Expected unsupported format error for %s, got: %v", ext, err)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(
		s, substr,
	)))
}

func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
