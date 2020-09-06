package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"unicode/utf8"

	"github.com/fogleman/gg"
)

type Config struct {
	BaseFile string `json:"base_file"`
	FontName string `json:"font_name"`
	FontSize int    `json:"font_size"`
	Bubbles  []struct {
		Id    int `json:"id"`
		Lines []struct {
			X      int `json:"x"`
			Y      int `json:"y"`
			Length int `json:"length"`
		} `json:"lines"`
	} `json:"bubbles"`
}

var config Config
var FontSize int

func toRotate(tc *gg.Context, s string, w float64, h float64) {
	switch s {
	case "ー", "〜", "～":
		tc.Rotate(gg.Radians(90))
		tc.DrawStringAnchored(s, h/2, 0-w/2, 0.5, 0.5)
		/*
		   case "。":
		     tc.Rotate(gg.Radians(90))
		     tc.DrawStringAnchored(s, h/2 + h/10, 0 - w, 0.5, 0.5)
		*/
	case "。", "、":
		tc.DrawString(s, w/2, h/3)
	default:
		tc.DrawStringAnchored(s, w/2, h/2, 0.5, 0.5)
	}
}

func measureString(s string) (float64, float64) {
	tempCampus := gg.NewContext(FontSize*2, FontSize*2)
	tempCampus.SetRGB(1, 1, 1)
	tempCampus.Clear()
	tempCampus.SetRGB(0, 0, 0)
	if err := tempCampus.LoadFontFace(config.FontName, float64(FontSize)); err != nil {
		panic(err)
	}
	w, h := tempCampus.MeasureString(s)
	return w, h
}

func createCharacterImage(s string) *gg.Context {
	w, h := measureString(s)
	w = w + (float64(FontSize) / 4)
	h = h + (float64(FontSize) / 4)
	tempCampus := gg.NewContext(int(w), int(h))
	tempCampus.SetRGB(1, 1, 1)
	tempCampus.Clear()
	tempCampus.SetRGB(0, 0, 0)
	if err := tempCampus.LoadFontFace(config.FontName, float64(FontSize)); err != nil {
		panic(err)
	}
	toRotate(tempCampus, s, w, h)
	return tempCampus
}

func PutStringInBubble(dc *gg.Context, msg string, x int, y int) {
	for _, c := range msg {
		s := string(c)
		ci := createCharacterImage(s)
		ci_size := ci.Image().Bounds().Size()
		dc.DrawImage(ci.Image(), x, y)
		y += ci_size.Y
	}
}

func main() {
	config_file := "config.json"
	quotes_file := "quotes.json"
	output_file := "output.png"
	if "" != os.Getenv("CONFIG_FILE") {
		config_file = os.Getenv("CONFIG_FILE")
	}
	if "" != os.Getenv("QUOTES_FILE") {
		quotes_file = os.Getenv("QUOTES_FILE")
	}
	if "" != os.Getenv("OUTPUT_FILE") {
		output_file = os.Getenv("OUTPUT_FILE")
	}
	bytes, err := ioutil.ReadFile(config_file)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatal(err)
	}
	ramen, err := gg.LoadPNG(config.BaseFile)
	if err != nil {
		panic(err)
	}
	ramen_size := ramen.Bounds().Size()
	dc := gg.NewContext(ramen_size.X, ramen_size.Y)
	dc.DrawImage(ramen, 0, 0)
	if err := dc.LoadFontFace(config.FontName, float64(FontSize)); err != nil {
		panic(err)
	}
	var quotes [][]string
	bytes, err = ioutil.ReadFile(quotes_file)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(bytes, &quotes); err != nil {
		log.Fatal(err)
	}
	for quote_index, lines := range quotes {
		minFontSize := config.FontSize
		for line_index, line := range lines {
			FontSize = config.FontSize
			length := config.Bubbles[quote_index].Lines[line_index].Length
			line_len := utf8.RuneCountInString(line)
			if line_len > length {
				delta := line_len - length
				if minFontSize > FontSize-(delta*2) {
					minFontSize = FontSize - (delta * 2)
				}
			}
		}
		FontSize = minFontSize
		for line_index, line := range lines {
			PutStringInBubble(dc, line, config.Bubbles[quote_index].Lines[line_index].X, config.Bubbles[quote_index].Lines[line_index].Y)
		}
	}
	dc.SavePNG(output_file)
}
