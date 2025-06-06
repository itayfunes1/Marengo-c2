package screenshot

import (
	"fmt"
	"image"
)

func ScreenSnapshot(fileName string) error {
	n := NumActiveDisplays()
	if n <= 0 {
		return fmt.Errorf("active display not found")
	}

	var all image.Rectangle = image.Rect(0, 0, 0, 0)

	for i := 0; i < n; i++ {
		bounds := GetDisplayBounds(i)
		all = bounds.Union(all)
	}
	img, err := Capture(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
	if err != nil {
		return fmt.Errorf("Cannot capture")
	}
	Savefile(fileName, img)
	return nil
}
