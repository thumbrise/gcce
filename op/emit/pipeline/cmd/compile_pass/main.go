package main

import (
	"log"
	"os"
	"path"

	"github.com/thumbrise/gcce/op/emit/pipeline/pass"
	"github.com/thumbrise/pipass"
)

const (
	outDir  = "op/emit/pipeline/pass"
	outFile = "pass.gen.go"
)

func main() {
	out, err := pipass.Compile(
		path.Base(outDir),
		pass.Instruction{},
		pass.Term{},
		pass.Operation{},
	)
	if err != nil {
		log.Fatalf("Compile: %v", err)
	}

	err = os.MkdirAll(outDir, 0o750)
	if err != nil {
		log.Fatalf("MkdirAll: %v", err)
	}

	err = os.WriteFile(path.Join(outDir, outFile), out, 0o600)
	if err != nil {
		log.Fatalf("WriteFile: %v", err)
	}
}
