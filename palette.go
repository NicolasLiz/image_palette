package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"

	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"charm.land/lipgloss/v2"
)


var help string = `command usage: palette <path_to_image> <number_of_colors>`

func absColor(value color.Color) int {
	r, g, b, _ := value.RGBA()

	return int(r + g + b)
}

func rgbaToHex(value color.Color) string {
	r, g, b, _ := value.RGBA()

	return strings.ToUpper(fmt.Sprintf("%02x%02x%02x", (r >> 8), (g >> 8), (b >> 8)))
}

func getImage(path string) (image.Image, image.Point) {
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	value, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	return value, value.Bounds().Size()
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal(help)
	}

	colors := make(map[color.Color]int)

	data, size := getImage(os.Args[1])
	for x := range size.X {
		for y := range size.Y {
			colors[data.At(x, y)]++
		}
	}

	resLen, _ := strconv.Atoi(os.Args[2])
	var res []struct {
		value  color.Color
		amount int
	}


	// color with highest total value (no alpha)
	var highest struct {
		value color.Color
		abs int
	}
	for k, _ := range colors {
		if absColor(k) > highest.abs {
			highest = struct{value color.Color; abs int}{k, absColor(k)}
		}
	}

	var style = lipgloss.NewStyle().
	Background(lipgloss.Color(fmt.Sprintf("#%s", rgbaToHex(highest.value))))
	lipgloss.Println(style.Render(fmt.Sprintf("#%s", rgbaToHex(highest.value))))

	// most occurring colors good for mono chromatic stuff
	for k, v := range colors {
		if len(res) < resLen {
			res = append(res, struct{value color.Color; amount int}{k, v})
			continue
		} else {
			for i, j := range res { 
				if v > j.amount {
					res[i] = struct{value color.Color; amount int}{k, v}		
					break
				}
			}
		}
	}


	for _, j := range res {
		var style = lipgloss.NewStyle().
		Background(lipgloss.Color(fmt.Sprintf("#%s", rgbaToHex(j.value))))
		lipgloss.Println(style.Render(fmt.Sprintf("#%s", rgbaToHex(j.value))))
	}

}
