package winutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows/registry"

	"github.com/TheTitanrain/w32"
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
	enumDisplayMonitors  = user32.NewProc("EnumDisplayMonitors")
)

// GetMonitors returns a list of monitors
func GetMonitors() ([]*Monitor, error) {
	// https://docs.microsoft.com/en-us/windows/desktop/api/Winuser/nf-winuser-enumdisplaymonitors
	out := []*Monitor{}
	ok := w32.EnumDisplayMonitors(
		w32.HDC(0),
		nil,
		syscall.NewCallback(func(hmon w32.HMONITOR, hdc w32.HDC, rect *w32.RECT, _ uintptr) uintptr {
			name, ok := getMonitorName(hmon)
			if !ok {
				name = ""
			}
			out = append(out, &Monitor{uintptr(hmon), uintptr(hdc), rect.Left, rect.Top, rect.Right, rect.Bottom, name})
			return 1
		}),
		uintptr(0),
	)
	if !ok {
		return out, errors.New("EnumDisplayMonitors was not successful")
	}
	return out, nil
}

func getMonitorName(hmon w32.HMONITOR) (string, bool) {
	// https://docs.microsoft.com/en-us/windows/desktop/api/Winuser/nf-winuser-getmonitorinfoa
	lmpi := &w32.MONITORINFOEX{}
	lmpi.CbSize = uint32(unsafe.Sizeof(lmpi))
	ok := w32.GetMonitorInfo(hmon, (*w32.MONITORINFO)(unsafe.Pointer(lmpi)))
	return syscall.UTF16ToString(lmpi.SzDevice[:]), ok
}

// Monitor represents a single display device
type Monitor struct {
	Hmonitor uintptr
	Hdc      uintptr
	Left     int32
	Top      int32
	Right    int32
	Bottom   int32
	Name     string
}

func GetWallpaper() (string, error) {
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

func SetWallpaper(path string) error {
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

func GetWallpaperStyle() (string, string, error) {
	style, err := getRegStr(registry.CURRENT_USER, `Control Panel\Desktop`, "WallpaperStyle")
	if err != nil {
		return "", "", err
	}
	tile, err := getRegStr(registry.CURRENT_USER, `Control Panel\Desktop`, "TileWallpaper")
	if err != nil {
		return "", "", err
	}
	return style, tile, nil
}

func SetWallpaperStyle(style string, tile string) error {
	// https://code.msdn.microsoft.com/windowsdesktop/CppSetDesktopWallpaper-eb969505
	err := setRegStr(registry.CURRENT_USER, `Control Panel\Desktop`, "WallpaperStyle", style)
	if err != nil {
		return err
	}
	err = setRegStr(registry.CURRENT_USER, `Control Panel\Desktop`, "TileWallpaper", tile)
	if err != nil {
		return err
	}
	return nil
}

func GetCurrentWallpaper(i int) (string, error) {
	value, err := readTranscodedImageCache(i)
	if value == "" && i > 0 {
		return GetCurrentWallpaper(0)
	}
	return value, err
}

func GetCurrentWallpapers() ([]string, error) {
	i := 0
	out := []string{}
	for {
		value, err := readTranscodedImageCache(i)
		if value == "" || err != nil {
			return out, err
		}
		out = append(out, value)
	}
}

func readTranscodedImageCache(i int) (string, error) {
	name := "TranscodedImageCache"
	withI := name + fmt.Sprintf("_%03d", i)

	if i > 0 {
		return _readTranscodedImageCacheName(withI)
	}

	value, err := _readTranscodedImageCacheName(withI)
	if err == registry.ErrNotExist {
		return _readTranscodedImageCacheName(name)
	}
	return value, err
}
func _readTranscodedImageCacheName(name string) (string, error) {
	bin, err := getRegBin(registry.CURRENT_USER, `Control Panel\Desktop`, name)
	if err == registry.ErrNotExist {
		return "", nil
	} else if err != nil {
		return "", err
	}

	bin = bin[24:]
	n := len(bin) / 2
	data := make([]byte, 0, n)
	for i := 0; i < n && bin[i] != 0; i += 2 {
		data = append(data, bin[i])
	}
	return string(data), nil
}

func getRegBin(root registry.Key, path string, name string) ([]byte, error) {
	key, err := registry.OpenKey(root, path, registry.READ)
	defer key.Close()
	if err != nil {
		return nil, err
	}
	bin, _, err := key.GetBinaryValue(name)
	return bin, err
}
func getRegStr(root registry.Key, path string, name string) (string, error) {
	key, err := registry.OpenKey(root, path, registry.READ)
	defer key.Close()
	if err != nil {
		return "", err
	}
	str, _, err := key.GetStringValue(name)
	return str, err
}
func setRegStr(root registry.Key, path string, name string, value string) error {
	key, err := registry.OpenKey(root, path, registry.READ|registry.WRITE)
	defer key.Close()
	if err != nil {
		return err
	}
	return key.SetStringValue(name, value)
}
