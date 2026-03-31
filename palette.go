package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"math"

	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"charm.land/lipgloss/v2"
)


var help string = `command usage: palette <path_to_image> <number_of_colors>`

func absColor(value color.Color) int {
	r, g, b, _ := value.RGBA()

	return int((r >> 8) + (g >> 8) + (b >> 8))
}

func rgbaToHex(value color.Color) string {
	r, g, b, _ := value.RGBA()

	return fmt.Sprintf("%02X%02X%02X", (r >> 8), (g >> 8), (b >> 8))
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
	for k := range colors {
		if absColor(k) > highest.abs && absColor(k) < 730 {
			highest = struct{value color.Color; abs int}{k, absColor(k)}
		}
	}

	var style = lipgloss.NewStyle().
	Background(lipgloss.Color(fmt.Sprintf("#%s", rgbaToHex(highest.value))))
	lipgloss.Println(style.Render(fmt.Sprintf("#%s", rgbaToHex(highest.value))))

	// color with lowest total value (no alpha)
	var lowest struct {
		value color.Color
		abs int
	}
	lowest.abs = highest.abs
	for k := range colors {
		if absColor(k) < lowest.abs && absColor(k) > 50 {
			lowest = struct{value color.Color; abs int}{k, absColor(k)}
		}
	}

	style = lipgloss.NewStyle().
	Background(lipgloss.Color(fmt.Sprintf("#%s", rgbaToHex(lowest.value))))
	lipgloss.Println(style.Render(fmt.Sprintf("#%s", rgbaToHex(lowest.value))))

	// most occurring colors good for mono chromatic stuff
	for k, v := range colors {
		if absColor(k) > absColor(highest.value) || absColor(k) < absColor(lowest.value) {
			continue
		}
		if len(res) < resLen {
			res = append(res, struct{value color.Color; amount int}{k, v})
			continue
		} else {
			for i, j := range res { 
				if v > j.amount {
					dif := 0
					for _, m := range res {
						if int(math.Abs(float64(absColor(k)) - float64(absColor(m.value)))) < 50 {
							dif = 1
							break
						}
					}
					if dif == 0 {
						res[i] = struct{value color.Color; amount int}{k, v}		
						break
					}
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
