package main

import (
	"fmt"
	"image"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
)

// loadBackgroundImage loads an image from the given path
func loadBackgroundImage(path string) (image.Image, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load background image: %w", err)
	}
	return img, nil
}

// resizeBackgroundImage resizes the image to match terminal dimensions
// Each cell in the terminal will correspond to one pixel in the resized image
func resizeBackgroundImage(imgInterface interface{}, width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return nil
	}

	// Type assert to image.Image
	img, ok := imgInterface.(image.Image)
	if !ok {
		return nil
	}

	// Use Lanczos resampling for better quality
	return imaging.Resize(img, width, height, imaging.Lanczos)
}

// renderBackgroundToString converts the image to a string of colored space characters
// Each pixel becomes a cell with that pixel's color as the background
func renderBackgroundToString(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var sb strings.Builder
	// Pre-allocate approximate capacity (each cell needs ~30 chars for ANSI codes)
	sb.Grow(width * height * 30)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert from 0-65535 range to 0-255
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			// Create background color from RGB
			bgColor := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r8, g8, b8))

			// Render a space with this background color
			style := lipgloss.NewStyle().Background(bgColor)
			sb.WriteString(style.Render(" "))
		}
		if y < height-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// getContrastColor calculates a contrasting foreground color for the given RGB background
// Uses relative luminance to determine if text should be black or white
func getContrastColor(r, g, b uint8) lipgloss.Color {
	// Calculate relative luminance using sRGB coefficients
	// https://www.w3.org/TR/WCAG20/#relativeluminancedef
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)

	// If background is bright, use dark text; otherwise use light text
	if lum > 128 {
		return lipgloss.Color("#000000") // Black
	}
	return lipgloss.Color("#ffffff") // White
}

// getAverageColor calculates the average color of an image region
// Useful for determining text color for a larger area
func getAverageColor(img image.Image, x, y, w, h int) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	x1 := max(x, bounds.Min.X)
	y1 := max(y, bounds.Min.Y)
	x2 := min(x+w, bounds.Max.X)
	y2 := min(y+h, bounds.Max.Y)

	var rSum, gSum, bSum uint64
	count := 0

	for py := y1; py < y2; py++ {
		for px := x1; px < x2; px++ {
			r, g, b, _ := img.At(px, py).RGBA()
			rSum += uint64(r >> 8)
			gSum += uint64(g >> 8)
			bSum += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return 128, 128, 128 // Default gray
	}

	return uint8(rSum / uint64(count)),
		uint8(gSum / uint64(count)),
		uint8(bSum / uint64(count))
}

// sampleRegionColor samples the average color from a region of the image
// Coordinates are given as percentages (0.0 to 1.0) of the image dimensions
func sampleRegionColor(img image.Image, xPercent, yPercent, wPercent, hPercent float64) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert percentages to pixel coordinates
	x := int(float64(width) * xPercent)
	y := int(float64(height) * yPercent)
	w := int(float64(width) * wPercent)
	h := int(float64(height) * hPercent)

	return getAverageColor(img, x, y, w, h)
}

// darkenColor darkens an RGB color by a factor (0.0 = black, 1.0 = unchanged)
func darkenColor(r, g, b uint8, factor float64) (uint8, uint8, uint8) {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return uint8(float64(r) * factor),
		uint8(float64(g) * factor),
		uint8(float64(b) * factor)
}

// brightenColor brightens an RGB color by a factor
func brightenColor(r, g, b uint8, factor float64) (uint8, uint8, uint8) {
	if factor < 1 {
		factor = 1
	}
	newR := int(float64(r) * factor)
	newG := int(float64(g) * factor)
	newB := int(float64(b) * factor)

	// Clamp to 255
	if newR > 255 {
		newR = 255
	}
	if newG > 255 {
		newG = 255
	}
	if newB > 255 {
		newB = 255
	}

	return uint8(newR), uint8(newG), uint8(newB)
}

// formatColor converts RGB to hex color string for lipgloss
func formatColor(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
