package main

import (
	"fmt"
)

func main() {
	t := Start(nil)

	t.Finish()

	// wait
	fmt.Scanln()
}