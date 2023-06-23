package commandui

import (
	"fmt"

	aic_package "github.com/TheZoraiz/ascii-image-converter/aic_package"
)

// printImageASCII request to print in terminal image from file.
func printImageASCII(path string) {
	flags := aic_package.DefaultFlags()
	flags.Colored = true

	answer, err := aic_package.Convert(path, flags)
	if err == nil {
		fmt.Printf("%s\n", answer)
	}
}

// return Flags{
// 	Complex:             false,
// 	Dimensions:          nil,
// 	Width:               0,
// 	Height:              0,
// 	SaveTxtPath:         "",
// 	SaveImagePath:       "",
// 	SaveGifPath:         "",
// 	Negative:            false,
// 	Colored:             false,
// 	CharBackgroundColor: false,
// 	Grayscale:           false,
// 	CustomMap:           "",
// 	FlipX:               false,
// 	FlipY:               false,
// 	Full:                false,
// 	FontFilePath:        "",
// 	FontColor:           [3]int{255, 255, 255},
// 	SaveBackgroundColor: [4]int{0, 0, 0, 100},
// 	Braille:             false,
// 	Threshold:           128,
// 	Dither:              false,
// 	OnlySave:            false,
