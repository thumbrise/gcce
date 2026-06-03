package mermaid

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

type MermaidBackend struct {
	graphDirection string
}

func NewMermaidBackend() *MermaidBackend {
	return &MermaidBackend{
		graphDirection: "TD",
	}
}

func (mb *MermaidBackend) Compile(operations []op.Operation) []byte {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "graph %s\n", mb.graphDirection)

	contractResolvers := make(map[string]string)

	buf.WriteString("    %% 1. Declaration of Composition Nodes\n")

	for _, operation := range operations {
		if len(operation.Output) == 0 {
			continue
		}

		target := operation.Output[0]
		targetID := sanitizeID(target.ID)
		displayTarget := displayName(target.ID)
		displayCtor := displayName(operation.ID)

		fmt.Fprintf(&buf, "    %s[\"🧩 %s<br><small>%s</small>\"]\n", targetID, displayTarget, displayCtor)

		for _, t := range target.Trait {
			if t.ID == trait.ImplementsID {
				contractFQN, ok := t.Value.(string)
				if !ok {
					continue
				}

				contractID := sanitizeID(contractFQN)
				displayContract := displayName(contractFQN)
				contractResolvers[contractID] = targetID

				fmt.Fprintf(&buf, "    %s((\"🔌 %s\"))\n", contractID, displayContract)
				fmt.Fprintf(&buf, "    %s -.->|fulfills| %s\n", targetID, contractID)
			}
		}
	}

	buf.WriteString("\n    %% 2. Execution and Dependency Routing\n")

	for _, operation := range operations {
		if len(operation.Output) == 0 {
			continue
		}

		targetID := sanitizeID(operation.Output[0].ID)

		for _, in := range operation.Input {
			depID := sanitizeID(in.ID)

			if contract, exists := contractResolvers[depID]; exists {
				fmt.Fprintf(&buf, "    %s --> %s\n", contract, targetID)
			} else {
				fmt.Fprintf(&buf, "    %s --> %s\n", depID, targetID)
			}
		}
	}

	return buf.Bytes()
}

func sanitizeID(fqn string) string {
	fqn = strings.ReplaceAll(fqn, "*", "")

	lastSlash := strings.LastIndex(fqn, "/")
	if lastSlash != -1 {
		fqn = fqn[lastSlash+1:]
	}

	return strings.ReplaceAll(fqn, ".", "_")
}

func displayName(fqn string) string {
	lastSlash := strings.LastIndex(fqn, "/")
	if lastSlash != -1 {
		return fqn[lastSlash+1:]
	}

	return fqn
}
