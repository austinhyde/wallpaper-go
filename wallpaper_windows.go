package wallpaper

import (
	"errors"
	"strconv"

	"github.com/austinhyde/wallpaper-go/winutils"
)

func getDesktopImpl() (Desktop, error) {
	monitors, err := winutils.GetMonitors()
	if err != nil {
		return nil, err
	}
	desktop := &winDesktop{make([]Screen, len(monitors))}
	for i, m := range monitors {
		desktop.screens[i] = &winScreen{desktop, m, i}
	}
	return desktop, nil
}

type winDesktop struct {
	screens []Screen // really *winScreen
}

// winDesktop implements Desktop

func (w *winDesktop) GetScreens() ([]Screen, error) {
	return w.screens, nil
}

// winDesktop implements Screen
// https://docs.microsoft.com/en-us/windows/desktop/gdi/the-virtual-screen

func (w *winDesktop) GetIdentifier() string {
	return "desktop"
}
func (w *winDesktop) GetWallpaper() (*Wallpaper, error) {
	wp, err := winutils.GetWallpaper()
	if err != nil {
		return nil, err
	}
	style, tile, err := winutils.GetWallpaperStyle()
	if err != nil {
		return nil, err
	}
	return &Wallpaper{wp, getStyleFromWin(style, tile)}, nil
}

func (w *winDesktop) SetWallpaper(wp *Wallpaper) error {
	return setWallpaperWithStyleImpl(wp.FilePath, wp.Style)
}

// winScreen implements Screen

type winScreen struct {
	desktop *winDesktop
	monitor *winutils.Monitor
	index   int
}

func (s *winScreen) GetIdentifier() string {
	if s.monitor.Name != "" {
		return s.monitor.Name
	}
	return strconv.Itoa(s.index)
}
func (s *winScreen) GetWallpaper() (*Wallpaper, error) {
	path, err := winutils.GetCurrentWallpaper(s.index)
	if err != nil {
		return nil, err
	}
	style, tile, err := winutils.GetWallpaperStyle()
	if err != nil {
		return nil, err
	}
	return &Wallpaper{path, getStyleFromWin(style, tile)}, nil
}
func (s *winScreen) SetWallpaper(*Wallpaper) error {
	return errors.New("not implemented")
}

// Helpers

func getWinStyle(s Style) string {
	//  WallpaperStyle
	//    0:  The image is centered if TileWallpaper=0 or tiled if TileWallpaper=1
	//    2:  The image is stretched to fill the screen
	//    6:  The image is resized to fit the screen while maintaining the aspect
	//        ratio. (Windows 7 and later)
	//    10: The image is resized and cropped to fill the screen while maintaining
	//        the aspect ratio. (Windows 7 and later)
	switch s {
	case StyleTile, StyleCenter:
		return "0"
	case StyleStretch:
		return "2"
	case StyleFit:
		return "6"
	}
	// StyleFill
	return "10"
}
func getWinTile(s Style) string {
	// TileWallpaper
	//    0: The wallpaper picture should not be tiled
	//    1: The wallpaper picture should be tiled
	if s == StyleTile {
		return "1"
	}
	return "0"
}
func getStyleFromWin(style string, tile string) Style {
	if tile == "1" {
		return StyleTile
	}
	switch style {
	case "0":
		return StyleCenter
	case "2":
		return StyleStretch
	case "6":
		return StyleFit
	}
	return StyleFill
}

func setWallpaperWithStyleImpl(path string, style Style) error {
	if style != StyleCurrent {
		if err := winutils.SetWallpaperStyle(getWinStyle(style), getWinTile(style)); err != nil {
			return err
		}
	}
	return winutils.SetWallpaper(path)
}

func setStyleImpl(style Style) error {
	// note: on windows, you need to re-set the current wallpaper for the style change to take effect
	curr, err := winutils.GetWallpaper()
	if err != nil {
		return err
	}
	return setWallpaperWithStyleImpl(curr, style)
}
