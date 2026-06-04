package main

import (
	"fmt"
	"log"

	"github.com/thumbrise/gcce/examples/opdsl/usecases"
)

func main() {
	h := usecases.NewHello("!")
	output, err := h.Hello("world")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
