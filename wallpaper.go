package wallpaper

import (
	"strings"
)

// GetDesktop returns an object representing the current desktop environment
func GetDesktop() (Desktop, error) {
	return getDesktopImpl()
}

// A Desktop is a collection of screens
type Desktop interface {
	GetScreens() ([]Screen, error)
}

// Screen objects can get/set a wallpaper
type Screen interface {
	GetIdentifier() string
	GetWallpaper() (*Wallpaper, error)
	SetWallpaper(*Wallpaper) error
}

// A Wallpaper has a path to a file, and a style that describes how it's laid out
type Wallpaper struct {
	FilePath string
	Style    Style
}

// A Rect is... a rectangle
type Rect struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

// GetScreenWithIdentifier returns the screen with the given identifier, if it exists
func GetScreenWithIdentifier(screens []Screen, id string) Screen {
	id = strings.TrimSpace(id)
	for _, s := range screens {
		if s.GetIdentifier() == id {
			return s
		}
	}
	return nil
}
