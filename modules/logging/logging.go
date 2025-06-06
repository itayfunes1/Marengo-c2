// logging.go
package logging

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	red   = color.New(color.FgRed).SprintFunc()
	green = color.New(color.FgGreen).SprintFunc()
	blue  = color.New(color.FgBlue).SprintFunc()
)

func logWithPrefix(prefix string, colorFunc func(...interface{}) string, args ...interface{}) {
	message := fmt.Sprint(args...)
	fmt.Printf("[%s] %s\n", colorFunc(prefix), message)
}

func Badln(args ...interface{}) {
	logWithPrefix("-", red, args...)
}

func Goodln(args ...interface{}) {
	logWithPrefix("+", green, args...)
}

func Infoln(args ...interface{}) {
	logWithPrefix("*", blue, args...)
}

func logWithFormatting(prefix string, colorFunc func(...interface{}) string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[%s] %s\n", colorFunc(prefix), message)
}

func Badf(format string, args ...interface{}) {
	logWithFormatting("-", red, format, args...)
}

func Goodf(format string, args ...interface{}) {
	logWithFormatting("+", green, format, args...)
}

func Infof(format string, args ...interface{}) {
	logWithFormatting("*", blue, format, args...)
}
