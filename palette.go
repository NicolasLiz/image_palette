package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"strconv"
	"sync"

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
	base     color.Color
	maxRange int
	parts    int
	lum      int
	cont     int
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

func partition(arr []colorField, low int, high int) int {
	pivot := arr[high]

	i := low - 1

	for j := low; j <= high-1; j++ {
		if arr[j].parts < pivot.parts {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func quickSortColors(arr []colorField, low int, high int) {
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
		case "--high": // highest luminosity
			options[opt] = os.Args[i+1]

		case "--low": // lowest luminosity
			options[opt] = os.Args[i+1]

		case "--lumdif":
			options[opt] = os.Args[i+1]

		case "--cntrdif":
			options[opt] = os.Args[i+1]

		case "--field": //field size
			options[opt] = os.Args[i+1]
		}

	}

	fieldSize := 10
	if val, ok := options["--field"]; ok {
		fieldSize, _ = strconv.Atoi(val)
	}

	maxLum := 765
	if val, ok := options["--high"]; ok {
		maxLum, _ = strconv.Atoi(val)
	}
	minLum := 0
	if val, ok := options["--low"]; ok {
		minLum, _ = strconv.Atoi(val)
	}

	lumDif := 0
	if val, ok := options["--lumdif"]; ok {
		lumDif, _ = strconv.Atoi(val)
	}

	contDif := 0
	if val, ok := options["--cntrdif"]; ok {
		contDif, _ = strconv.Atoi(val)
	}

	maxColors, err := strconv.Atoi(numColors)
	if err != nil {
		log.Fatal(err)
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
	var colorFields []*colorField
	colorFields = append(colorFields, &colorField{imageData.At(0, 0), 0, 1, 0, 0})

	var wg sync.WaitGroup
	for x := range imageSize.X {
		wg.Go(func() {
			for y := range imageSize.Y {
				for i, k := range colorFields {
					data := imageData.At(x, y)
					if inField((*k), data) {
						k.parts++
						break
					} else if i == len(colorFields)-1 {
						colorFields = append(colorFields, &colorField{data, fieldSize, 1, luminosity(data), 0})
					}
				}
			}
		})
	}
	wg.Wait()

	colorFields = colorFields[1:]

	// color with highest and lowest luminosity
	var highest colorField
	var lowest colorField

	for _, v := range colorFields {
		//highest lum && cont
		if highest.parts == 0 {
			highest = (*v)
			highest.cont = 441
		}
		if v.lum >= highest.lum && v.lum <= maxLum && contrast(v.base, color.White) < highest.cont {
			highest = (*v)
			highest.cont = contrast(v.base, color.White)
		}

		//lowest lum && cont
		if lowest.parts == 0 {
			lowest = (*v)
			highest.cont = 441
		}
		if v.lum <= lowest.lum && v.lum >= minLum && contrast(v.base, color.Black) < lowest.cont {
			lowest = (*v)
			lowest.cont = contrast(v.base, color.Black)
		}
	}

	var result []colorField

	for _, v := range colorFields {
		if v.lum > highest.lum || v.lum < lowest.lum {
			continue
		}

		if (*v) == highest || (*v) == lowest {
			continue
		}

		f := 0
		for _, j := range result {
			if int(math.Abs(float64(v.lum)-float64(j.lum))) < lumDif {
				f = 1
				break
			}
			if contrast(v.base, j.base) < contDif {
				f = 1
				break
			}
		}
		if f == 1 {
			continue
		}

		result = append(result, (*v))
	}

	quickSortColors(result, 0, len(result) - 1)

	fmt.Println("brightest")
	printColor(highest.base)
	fmt.Println("darkest")
	printColor(lowest.base)

	if len(result) > maxColors {
		result = result[len(result)-maxColors:]
	}
	for _, j := range result {
		printColor(j.base)
	}
}
