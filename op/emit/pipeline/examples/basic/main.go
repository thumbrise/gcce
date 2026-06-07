package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/thumbrise/gcce/op/emit/pipeline"
	"github.com/thumbrise/gcce/op/emit/pipeline/contract"
	"github.com/thumbrise/gcce/op/emit/pipeline/examples/basic/usecases"
)

func main() {
	h := usecases.NewHello("!")
	ppln := pipeline.NewPipeline(
		contract.InstructionRegistration{
			ID:      "hello-program",
			Version: "v0.1.0",
			Comment: "Hello World example",
		},
		[]contract.OperationRegistration{
			{FN: h.Hello},
		},
	)

	instr, err := ppln.Compile()
	if err != nil {
		log.Fatalf("pipeline.Compile(): %v", err)
	}

	j, err := json.MarshalIndent(instr, "", "\t")
	if err != nil {
		log.Fatalf("json.MarshalIndent: %v", err)
	}
	fmt.Println(string(j))
}
