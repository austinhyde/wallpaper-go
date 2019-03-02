package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/austinhyde/wallpaper-go"
)

func main() {
	var styleStr, monitor string
	var listMonitorsFlag bool
	flag.BoolVar(&listMonitorsFlag, "list", false, "print the list of screens and their current wallpaper")
	flag.StringVar(&styleStr, "style", "", "fill style to use when setting a wallpaper: current[default],fill,fit,stretch,center,tile")
	flag.StringVar(&monitor, "screen", "", "screen identifier to set (use -list to find available ones)")
	flag.Parse()

	if listMonitorsFlag {
		listMonitors()
		os.Exit(0)
	}

	style := wallpaper.ParseStyleString(styleStr)

	if flag.NArg() == 0 {
		if style != wallpaper.StyleCurrent {
			setStyle(style, monitor)
		} else {
			getCurrent(monitor)
		}
	} else {
		setCurrent(flag.Arg(0), style, monitor)
	}
}

func listMonitors() {
	desktop, err := wallpaper.GetDesktop()
	checkErr("Could not list screens", err)

	screens, err := desktop.GetScreens()
	checkErr("Could not list screens", err)
	for i, screen := range screens {
		ident := screen.GetIdentifier()
		if ident == "" {
			ident = strconv.Itoa(i)
		}
		wp, err := screen.GetWallpaper()

		if err != nil {
			fmt.Printf("%s  <could not get wallpaper:%s>\n", ident, err)
		} else {
			fmt.Printf("%s  %s\n", ident, wp.FilePath)
		}
	}
}

func getCurrent(id string) {
	wp, err := getScreen(id).GetWallpaper()
	checkErr("Could not get wallpaper path", err)
	fmt.Println(wp.FilePath)
}
func setCurrent(path string, style wallpaper.Style, id string) {
	err := getScreen(id).SetWallpaper(&wallpaper.Wallpaper{path, style})
	checkErr("Could not set wallpaper", err)
}

func setStyle(style wallpaper.Style, id string) {
	screen := getScreen(id)
	curr, err := screen.GetWallpaper()
	checkErr("Could not set wallpaper style", err)

	err = screen.SetWallpaper(&wallpaper.Wallpaper{curr.FilePath, style})
	checkErr("Could not set wallpaper style", err)
}

func getScreen(id string) wallpaper.Screen {
	desktop, err := wallpaper.GetDesktop()
	checkErr("Could not get desktop info", err)

	if id != "" {
		screens, err := desktop.GetScreens()
		checkErr("Could not enumerate screens", err)

		screen := wallpaper.GetScreenWithIdentifier(screens, id)
		if screen == nil {
			fmt.Printf("Screen '%s' was not found\n", id)
			os.Exit(1)
		}
		return screen
	} else if virtual, ok := desktop.(wallpaper.Screen); ok {
		return virtual
	} else {
		fmt.Println("A monitor identifier must be specified")
		os.Exit(1)
		return nil
	}
}

func checkErr(msg string, err error) {
	if err != nil {
		fmt.Printf(msg+": %s\n", err)
		os.Exit(1)
	}
}
