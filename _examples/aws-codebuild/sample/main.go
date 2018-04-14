package main

import (
	"fmt"
)

var (
	// buildVersion and buildDate are populated with values during compile-time
	buildVersion string
	buildDate    string
)

func main() {
	fmt.Println("Hello World")
}
