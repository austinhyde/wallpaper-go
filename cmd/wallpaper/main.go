package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/austinhyde/wallpaper-go"
)

func main() {
	var styleStr string
	flag.StringVar(&styleStr, "style", "", "fill style to use when setting a wallpaper: current[default],fill,fit,stretch,center,tile")
	flag.Parse()

	style := wallpaper.ParseStyleString(styleStr)

	if flag.NArg() == 0 {
		if style != wallpaper.StyleCurrent {
			setStyle(style)
		} else {
			getCurrent()
		}
	} else {
		setCurrent(flag.Arg(0), style)
	}
}

func getCurrent() {
	path, err := wallpaper.GetWallpaper()
	if err != nil {
		fmt.Printf("Could not get wallpaper path: %s\n", err)
		os.Exit(1)
	} else {
		fmt.Println(path)
	}
}
func setCurrent(path string, style wallpaper.Style) {
	err := wallpaper.SetWallpaperWithStyle(flag.Arg(0), style)
	if err != nil {
		fmt.Printf("Could not set wallpaper: %s\n", err)
		os.Exit(1)
	}
}
func setStyle(style wallpaper.Style) {
	err := wallpaper.SetStyle(style)
	if err != nil {
		fmt.Printf("Could not set wallpaper: %s\n", err)
		os.Exit(1)
	}
}
