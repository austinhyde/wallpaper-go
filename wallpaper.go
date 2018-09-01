package wallpaper

import (
	"strings"
)

// Style denotes how a wallpaper is sized on the screen
type Style string

const (
	// StyleCurrent uses the currently configured wallpaper style
	StyleCurrent Style = ""
	// StyleFill sets the wallpaper to fill available space while keeping aspect ratio
	StyleFill = "fill"
	// StyleFit sets the wallpaper to fit in available space while keeping aspect ratio
	StyleFit = "fit"
	// StyleStretch sets the wallpaper to deform to fill available space
	StyleStretch = "stretch"
	// StyleCenter centers the wallpaper on the desktop without changing its size
	StyleCenter = "center"
	// StyleTile repeats the wallpaper horizontally and vertically to fill available space
	StyleTile = "tile"
)

// ParseStyleString interprets a string to be a Style
func ParseStyleString(s string) Style {
	if s == "current" {
		s = ""
	}
	return Style(strings.ToLower(s))
}

// IsValid returns true if the style is a known value
func (s Style) IsValid() bool {
	switch s {
	case StyleCurrent, StyleFill, StyleFit, StyleStretch, StyleCenter, StyleTile:
		return true
	}
	return false
}

// GetWallpaper returns the filesystem path to the current wallpaper image
func GetWallpaper() (string, error) {
	return getWallpaperImpl()
}

// SetWallpaper sets the current wallpaper to the image
func SetWallpaper(path string) error {
	return setWallpaperImpl(path)
}

// SetWallpaperWithStyle sets the current wallpaper to the image at the given path with the given style
func SetWallpaperWithStyle(path string, style Style) error {
	return setWallpaperWithStyleImpl(path, style)
}

// SetStyle sets the current wallpaper style without changing the image
func SetStyle(style Style) error {
	return setStyleImpl(style)
}
