package pipeline

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/thumbrise/gcce/op/emit"
	"github.com/thumbrise/gcce/op/emit/pipeline/contract"
	passgen "github.com/thumbrise/gcce/op/emit/pipeline/pass"
	"github.com/thumbrise/gcce/op/emit/pipeline/util"
	"github.com/thumbrise/op-universal-schema-go/schema"
	"github.com/thumbrise/pipass"
)

var (
	ErrPluginFailure = errors.New("plugin failure")
	ErrEmitFailure   = errors.New("emit failure")
)

const (
	coreEnrichmentReason      = "core enrichment"
	userDefineReason          = "user define"
	instructionVersionUnknown = "v0.0.0-unknown"
	instructionIDUnknown      = "unknown"
	instructionCommentUnknown = "unknown"
)

type Pipeline struct {
	instructionRegistration contract.InstructionRegistration
	operationRegistrations  []contract.OperationRegistration
	Ledger                  pipass.Ledger
}

func NewPipeline(instructionRegistration contract.InstructionRegistration, operationRegistrations []contract.OperationRegistration) *Pipeline {
	return &Pipeline{
		instructionRegistration: instructionRegistration,
		operationRegistrations:  operationRegistrations,
	}
}

func (p *Pipeline) Compile() (*schema.Instruction, error) {
	ledger := p.Ledger
	if ledger == nil {
		ledger = NewPrettyDiffLedger(os.Stdout)
	}

	instructionPass := passgen.NewInstructionPipePass("instruction", ledger)
	setInstructionMetadata(instructionPass, p.instructionRegistration)

	for _, registration := range p.operationRegistrations {
		opPass, err := buildOperationPass(registration.FN, instructionPass, ledger)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEmitFailure, err)
		}

		for _, plugin := range registration.Plugins {
			if err := plugin.Visit(opPass); err != nil {
				return nil, fmt.Errorf("%w: plugin %s: %w", ErrPluginFailure, plugin.Name(), err)
			}
		}
	}

	for _, plugin := range p.instructionRegistration.Plugins {
		if err := plugin.Visit(instructionPass); err != nil {
			return nil, fmt.Errorf("%w: plugin %s: %w", ErrPluginFailure, plugin.Name(), err)
		}
	}

	return instructionToSchema(instructionPass), nil
}

func instructionToSchema(instrPass *passgen.InstructionPipePass) *schema.Instruction {
	if instrPass.Dropped() {
		return nil
	}

	result := &schema.Instruction{
		ID:      instrPass.ID(),
		Comment: instrPass.Comment(),
		Version: instrPass.Version(),
	}

	for _, op := range instrPass.Operations() {
		if op.Dropped() {
			continue
		}

		result.Operations = append(result.Operations, toSchemaOperation(op))
	}

	for _, t := range instrPass.Trait() {
		if t.Dropped() {
			continue
		}

		result.Trait = append(result.Trait, toSchemaTerm(t))
	}

	return result
}

func setInstructionMetadata(instructionPass *passgen.InstructionPipePass, reg contract.InstructionRegistration) {
	if reg.ID != "" {
		instructionPass.SetID(reg.ID, userDefineReason)
	} else {
		instructionPass.SetID(instructionIDUnknown, coreEnrichmentReason)
	}

	if reg.Comment != "" {
		instructionPass.SetComment(reg.Comment, userDefineReason)
	} else {
		instructionPass.SetComment(instructionCommentUnknown, coreEnrichmentReason)
	}

	if reg.Version != "" {
		instructionPass.SetVersion(reg.Version, userDefineReason)
	} else {
		instructionPass.SetVersion(instructionVersionUnknown, coreEnrichmentReason)
	}
}

func buildOperationPass(fn interface{}, instrPass *passgen.InstructionPipePass, ledger pipass.Ledger) (*passgen.OperationPipePass, error) {
	operation, err := emit.FunctionToOperation(fn)
	if err != nil {
		return nil, err
	}

	opPass := passgen.NewOperationPipePass("", ledger)
	opPass.SetID(operation.ID, coreEnrichmentReason)
	opPass.SetComment(operation.Comment, coreEnrichmentReason)
	opPass.SetReflectType(reflect.TypeOf(fn), coreEnrichmentReason)

	fd, err := util.FuncDecl(fn)
	if err == nil {
		opPass.SetASTFuncDecl(fd, coreEnrichmentReason)
	}

	for _, input := range operation.Input {
		opPass.AppendInput(termToPass(input, ledger), coreEnrichmentReason)
	}

	for _, output := range operation.Output {
		opPass.AppendOutput(termToPass(output, ledger), coreEnrichmentReason)
	}

	for _, errTerm := range operation.Error {
		opPass.AppendError(termToPass(errTerm, ledger), coreEnrichmentReason)
	}

	for _, traitTerm := range operation.Trait {
		opPass.AppendTrait(termToPass(traitTerm, ledger), coreEnrichmentReason)
	}

	instrPass.AppendOperations(opPass, coreEnrichmentReason)

	return opPass, nil
}

func termToPass(t schema.Term, ledger pipass.Ledger) *passgen.TermPipePass {
	tp := passgen.NewTermPipePass("", ledger)
	tp.SetID(t.ID, coreEnrichmentReason)
	tp.SetComment(t.Comment, coreEnrichmentReason)
	tp.SetRequired(t.Required, coreEnrichmentReason)
	tp.SetKind(t.Kind, coreEnrichmentReason)
	tp.SetValue(t.Value, coreEnrichmentReason)

	for _, of := range t.Of {
		tp.AppendOf(termToPass(of, ledger), coreEnrichmentReason)
	}

	for _, tr := range t.Trait {
		tp.AppendTrait(termToPass(tr, ledger), coreEnrichmentReason)
	}

	return tp
}

func toSchemaTerm(tp passgen.TermPass) schema.Term {
	result := schema.Term{
		ID:       tp.ID(),
		Comment:  tp.Comment(),
		Required: tp.Required(),
		Kind:     tp.Kind(),
		Value:    tp.Value(),
	}

	for _, of := range tp.Of() {
		if of.Dropped() {
			continue
		}

		result.Of = append(result.Of, toSchemaTerm(of))
	}

	for _, tr := range tp.Trait() {
		if tr.Dropped() {
			continue
		}

		result.Trait = append(result.Trait, toSchemaTerm(tr))
	}

	return result
}

func toSchemaOperation(op passgen.OperationPass) schema.Operation {
	result := schema.Operation{
		ID:      op.ID(),
		Comment: op.Comment(),
	}

	for _, input := range op.Input() {
		if input.Dropped() {
			continue
		}

		result.Input = append(result.Input, toSchemaTerm(input))
	}

	for _, output := range op.Output() {
		if output.Dropped() {
			continue
		}

		result.Output = append(result.Output, toSchemaTerm(output))
	}

	for _, errTerm := range op.Error() {
		if errTerm.Dropped() {
			continue
		}

		result.Error = append(result.Error, toSchemaTerm(errTerm))
	}

	for _, tr := range op.Trait() {
		if tr.Dropped() {
			continue
		}

		result.Trait = append(result.Trait, toSchemaTerm(tr))
	}

	return result
}
