// +build darwin

package main

import (
	"fmt"
	"ziphttp"
)

func ColorRed(s string) {
	fmt.Println(ziphttp.ColorRed(s))
}

func ColorGreen(s string) {
	fmt.Println(ziphttp.ColorGreen(s))
}

func ColorBlue(s string) {
	fmt.Println(ziphttp.ColorBlue(s))
}
