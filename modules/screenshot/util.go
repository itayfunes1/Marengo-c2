package screenshot

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
)

func CreateImage(rect image.Rectangle) (img *image.RGBA, e error) {
	img = nil
	e = errors.New("Cannot create image.RGBA")

	defer func() {
		err := recover()
		if err == nil {
			e = nil
		}
	}()
	// image.NewRGBA may panic if rect is too large.
	img = image.NewRGBA(rect)

	return img, e
}

func Savefile(fileName string, img *image.RGBA) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err = png.Encode(file, img); err != nil {
		fmt.Println("ERROR SAVING IMAGE: ", err)
	}
}
