package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/thumbrise/gcce/examples/opdsl/usecases"
	"github.com/thumbrise/gcce/op/emit/struc"
)

func main() {
	ops, err := struc.T(new(usecases.Hello))
	if err != nil {
		log.Fatalf("struc.T: %v", err)
	}
	j, err := json.MarshalIndent(ops, "", "\t")
	if err != nil {
		log.Fatalf("json.MarshalIndent: %v", err)
	}
	fmt.Println(string(j))
}
