package wallpaper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// Adapted from https://github.com/reujab/wallpaper/blob/master/windows.go
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724947.aspx
const (
	spiGetDeskWallpaper = 0x0073
	spiSetDeskWallpaper = 0x0014

	uiParam = 0x0000

	spifUpdateINIFile = 0x01
	spifSendChange    = 0x02
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

func getWallpaperImpl() (string, error) {
	var filename [256]uint16
	_, _, err := systemParametersInfo.Call(
		uintptr(spiGetDeskWallpaper),
		uintptr(cap(filename)),
		// the memory address of the first byte of the array
		uintptr(unsafe.Pointer(&filename[0])),
		uintptr(0),
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		return "", err
	}
	return strings.Trim(string(utf16.Decode(filename[:])), "\x00"), nil
}

func setWallpaperImpl(path string) error {
	return setWallpaperWithStyleImpl(path, StyleCurrent)
}

func setWallpaperWithStyleImpl(path string, style Style) error {
	if style != StyleCurrent {
		if err := setStyle(style); err != nil {
			return err
		}
	}
	return setWallpaper(path)
}

func setStyleImpl(style Style) error {
	// note: on windows, you need to re-set the current wallpaper for the style change to take effect
	curr, err := GetWallpaper()
	if err != nil {
		return err
	}
	if err = setStyle(style); err != nil {
		return err
	}
	return setWallpaper(curr)
}

func setWallpaper(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if err != nil {
		return err
	}
	filenameUTF16, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	_, _, err = systemParametersInfo.Call(
		uintptr(spiSetDeskWallpaper),
		uintptr(uiParam),
		uintptr(unsafe.Pointer(filenameUTF16)),
		uintptr(spifUpdateINIFile|spifSendChange),
	)
	if err != nil {
		switch err.Error() {
		case "The operation completed successfully.", "This operation returned because the timeout period expired.":
			return nil
		}
		return err
	}
	return nil
}

func setStyle(style Style) error {
	if !style.IsValid() {
		return fmt.Errorf("Invalid style parameter: '%s'", style)
	}
	// this is a no-op
	if style == StyleCurrent {
		return nil
	}

	// https://code.msdn.microsoft.com/windowsdesktop/CppSetDesktopWallpaper-eb969505
	key, err := registry.OpenKey(registry.CURRENT_USER, "Control Panel\\Desktop", registry.READ|registry.WRITE)
	defer key.Close()
	if err != nil {
		return err
	}

	err = key.SetExpandStringValue("WallpaperStyle", style.getWinStyle())
	if err != nil {
		return err
	}
	err = key.SetExpandStringValue("TileWallpaper", style.getWinTile())
	if err != nil {
		return err
	}
	return nil
}

func (s Style) getWinStyle() string {
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
func (s Style) getWinTile() string {
	// TileWallpaper
	//    0: The wallpaper picture should not be tiled
	//    1: The wallpaper picture should be tiled
	if s == StyleTile {
		return "1"
	}
	return "0"
}
