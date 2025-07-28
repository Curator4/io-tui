package visual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cascax/colorthief-go"
	"github.com/qeesung/image2ascii/convert"
)

// GenerateFromImageURL downloads an image from URL and generates both
// a color palette and ASCII art from it
func GenerateFromImageURL(imageURL string) (palette []string, ascii string, err error) {
	// Download image to temporary file
	tempPath, err := downloadImage(imageURL)
	if err != nil {
		return nil, "", fmt.Errorf("❌ Failed to download image from URL")
	}
	defer os.Remove(tempPath) // Clean up temp file
	
	// Extract color palette
	palette, err = extractPalette(tempPath)
	if err != nil {
		return nil, "", fmt.Errorf("🎨 Failed to extract color palette from image")
	}
	
	// Generate ASCII art
	ascii, err = generateASCII(tempPath)
	if err != nil {
		return nil, "", fmt.Errorf("🖼️ Failed to generate ASCII art from image")
	}
	
	return palette, ascii, nil
}

// extractPalette uses colorthief to extract dominant colors from an image
func extractPalette(imagePath string) ([]string, error) {
	// Extract 8 dominant colors
	colors, err := colorthief.GetPaletteFromFile(imagePath, 8)
	if err != nil {
		return nil, fmt.Errorf("failed to extract palette: %w", err)
	}
	
	// Convert to hex strings
	var palette []string
	for _, color := range colors {
		r, g, b, _ := color.RGBA()
		// Convert from 16-bit to 8-bit values
		hex := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
		palette = append(palette, hex)
	}
	
	return palette, nil
}

// generateASCII converts an image to ASCII art
func generateASCII(imagePath string) (string, error) {
	// Create converter with options
	converter := convert.NewImageConverter()
	
	// Set options for colored ASCII output
	options := convert.DefaultOptions
	options.FixedWidth = 30    // Fixed width for consistent display
	options.FixedHeight = 20   // Fixed height for header area
	options.Colored = true     // Enable ANSI color codes in ASCII
	
	// Convert image to ASCII
	ascii := converter.ImageFile2ASCIIString(imagePath, &options)
	
	// Ensure exactly 20 lines (preserve ANSI escape sequences)
	ascii = strings.TrimSpace(ascii)
	lines := strings.Split(ascii, "\n")
	if len(lines) > 20 {
		// Only trim if we have more than 20 actual lines
		lines = lines[:20]
		ascii = strings.Join(lines, "\n")
	}
	
	return ascii, nil
}

// downloadImage downloads an image from URL to a temporary file
func downloadImage(url string) (string, error) {
	// Create HTTP client with timeout and proper headers
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Create request with proper headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add user agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; io-tui/1.0)")
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}
	
	// Create temporary file
	tempFile, err := os.CreateTemp("", "ai-image-*.jpg")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	
	// Copy response body to file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name()) // Clean up on error
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	
	return tempFile.Name(), nil
}

// FormatPaletteForDB converts a palette slice to JSON string for database storage
func FormatPaletteForDB(palette []string) (string, error) {
	data, err := json.Marshal(palette)
	if err != nil {
		return "", fmt.Errorf("failed to marshal palette: %w", err)
	}
	return string(data), nil
}

// ParsePaletteFromDB converts JSON string from database back to palette slice
func ParsePaletteFromDB(paletteJSON string) ([]string, error) {
	var palette []string
	if err := json.Unmarshal([]byte(paletteJSON), &palette); err != nil {
		return nil, fmt.Errorf("failed to unmarshal palette: %w", err)
	}
	return palette, nil
}
