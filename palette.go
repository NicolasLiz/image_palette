package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"strconv"

	"image/color"
	_ "image/jpeg"
	_ "image/png"

	"charm.land/lipgloss/v2"
)

var help string = `command usage: palette <path_to_image> <number_of_colors> [options]`

type colorData struct {
	value  color.Color
	amount int
	abs    int
	cont   int
}

type colorField struct {
	base color.Color
	maxRange int
	parts int
}

func inField(field colorField, elem color.Color) bool {
	if contrast(field.base, elem) <= field.maxRange {
		return true
	}
	return false
}

func avrgColor(arr []colorData) colorData {
	var r, g, b, l, c int
	for _, v := range arr {
		l += v.abs
		c += v.cont
		tr, tg, tb, _ := v.value.RGBA()
		r += int(tr)
		g += int(tg)
		b += int(tb)
	}
	r /= len(arr)
	g /= len(arr)
	b /= len(arr)
	r = (r >> 8)
	g = (g >> 8)
	b = (b >> 8)
	l /= len(arr)
	c /= len(arr)
	return colorData{color.RGBA{uint8(r), uint8(g), uint8(b), 0}, 0, l, c}
}

func popHighest(colors map[color.Color]int) color.Color {
	var res color.Color
	high := 0

	for i, j := range colors {
		if j > high {
			res = i
			high = j
		}
	}

	delete(colors, res)

	return res
}

func contrast(color1 color.Color, color2 color.Color) int {
	r, g, b, _ := color1.RGBA()
	r2, g2, b2, _ := color2.RGBA()

	r = (r >> 8)
	g = (g >> 8)
	b = (b >> 8)

	r2 = (r2 >> 8)
	g2 = (g2 >> 8)
	b2 = (b2 >> 8)

	r2 -= r
	g2 -= g
	b2 -= b

	r2 *= r2
	g2 *= g2
	b2 *= b2

	return int(math.Abs(math.Sqrt(float64(r2) + float64(g2) + float64(b2))))
}

func partition(arr []colorData, low int, high int) int {
	pivot := arr[high]

	i := low - 1

	for j := low; j <= high-1; j++ {
		if arr[j].amount < pivot.amount {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func quickSortColors(arr []colorData, low int, high int) {
	if low < high {
		pi := partition(arr, low, high)

		quickSortColors(arr, low, pi-1)
		quickSortColors(arr, pi+1, high)
	}
}

func printColor(clr color.Color) {
	var style = lipgloss.NewStyle().
		Background(lipgloss.Color(fmt.Sprintf("#%s", rgbaToHex(clr))))
	lipgloss.Println(style.Render(fmt.Sprintf("#%s", rgbaToHex(clr))))
}

func luminosity(value color.Color) int {
	r, g, b, _ := value.RGBA()

	return int((r >> 8) + (g >> 8) + (b >> 8))
}

func rgbaToHex(value color.Color) string {
	r, g, b, _ := value.RGBA()

	return fmt.Sprintf("%02X%02X%02X", (r >> 8), (g >> 8), (b >> 8))
}

func main() {
	// process args
	if len(os.Args) < 3 {
		log.Fatal(help)
	}

	imagePath := os.Args[1]
	numColors := os.Args[2]

	options := make(map[string]string)

	for i, opt := range os.Args {
		switch opt {
		case "-h": // highest luminosity
			options[opt] = os.Args[i+1]

		case "-l": // lowest luminosity
			options[opt] = os.Args[i+1]

		case "-d":
			options[opt] = os.Args[i+1]

		case "-c":
			options[opt] = os.Args[i+1]
		}
	}

	// opens image
	f, err := os.Open(imagePath)
	if err != nil {
		log.Fatal(err)
	}

	imageData, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	imageSize := imageData.Bounds().Size()

	f.Close()

	// creates map of colors and how many time they appear
	colors := make(map[color.Color]int)
	var colorFields []colorField
	colorFields = append(colorFields, colorField{imageData.At(0, 0), 30, 1})

	for x := range imageSize.X {
		for y := range imageSize.Y {
			colors[imageData.At(x, y)]++
			for i, k := range colorFields {
				if inField(k, imageData.At(x, y)) {
					k.parts++
					break
				} else if i == len(colorFields) - 1 {
					colorFields = append(colorFields, colorField{imageData.At(x, y), 30, 1})
				}
			}
		}
	}

	// color with highest and lowest luminosity
	var highest colorData
	var lowest colorData
	lowest.abs = 766

	maxLum := 766
	if val, ok := options["-h"]; ok {
		maxLum, _ = strconv.Atoi(val)
	}
	minLum := 0
	if val, ok := options["-l"]; ok {
		minLum, _ = strconv.Atoi(val)
	}

	for k, v := range colors {
		//highest lum && cont
		lumK := luminosity(k)
		if highest.cont == 0 {
			highest = colorData{k, v, luminosity(k), contrast(k, color.White)}
		}
		if lumK >= highest.abs && lumK <= maxLum && contrast(k, color.White) < highest.cont {
			highest = colorData{k, v, luminosity(k), contrast(k, color.White)}
		}

		//lowest lum && cont
		if lowest.cont == 0 {
			lowest = colorData{k, v, luminosity(k), contrast(k, color.Black)}
		}
		if lumK <= lowest.abs && lumK >= minLum && contrast(k, color.Black) < lowest.cont {
			lowest = colorData{k, v, luminosity(k), contrast(k, color.Black)}
		}
	}

	maxColors, err := strconv.Atoi(numColors)
	if err != nil {
		log.Fatal(err)
	}

	lumDif := 0
	if val, ok := options["-d"]; ok {
		lumDif, _ = strconv.Atoi(val)
	}

	contDif := 0
	if val, ok := options["-c"]; ok {
		contDif, _ = strconv.Atoi(val)
	}

	var result []colorData

	for k, v := range colors {
		lumK := luminosity(k)

		if lumK > highest.abs || lumK < lowest.abs {
			continue
		}

		f := 0
		for _, j := range result {
			if int(math.Abs(float64(lumK)-float64(j.abs))) < lumDif {
				f = 1
				break
			}
			if contrast(k, j.value) < contDif {
				f = 1
				break
			}
		}
		if f == 1 {
			continue
		}

		result = append(result, colorData{k, v, 0, 0})
	}

	quickSortColors(result, 0, len(result)-1)

	printColor(highest.value)
	printColor(lowest.value)

	if len(result) > maxColors {
		result = result[len(result)-maxColors:]
	}
	for _, j := range result {
		printColor(j.value)
	}
}
