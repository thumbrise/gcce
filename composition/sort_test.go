package composition_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/composition"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
	op "github.com/thumbrise/op-universal-schema-go/schema"
)

func opStep(id, target string, deps ...string) op.Operation {
	in := make([]op.Term, len(deps))
	for i, d := range deps {
		in[i] = op.Term{ID: d}
	}

	return op.Operation{
		ID:     id,
		Output: []op.Term{{ID: target}},
		Input:  in,
	}
}

func opWithImplements(id, target string, contracts []string, deps ...string) op.Operation {
	input := make([]op.Term, len(deps))
	for i, d := range deps {
		input[i] = op.Term{ID: d}
	}

	outputTargetTraits := make([]op.Term, 0, len(contracts))
	for _, contract := range contracts {
		outputTargetTraits = append(outputTargetTraits, trait.NewImplements(contract))
	}

	output := []op.Term{{
		ID:    target,
		Trait: outputTargetTraits,
	}}

	return op.Operation{
		ID:     id,
		Output: output,
		Input:  input,
	}
}

func TestSort_NoDeps(t *testing.T) {
	a := opStep("pkg.A", "A")
	b := opStep("pkg.B", "B")
	c := opStep("pkg.C", "C")

	got, err := composition.SortOperations([]op.Operation{a, b, c})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "pkg.A", got[0].ID)
	require.Equal(t, "pkg.B", got[1].ID)
	require.Equal(t, "pkg.C", got[2].ID)
}

func TestSort_LinearChain(t *testing.T) {
	a := opStep("pkg.A", "A")
	b := opStep("pkg.B", "B", "A")
	c := opStep("pkg.C", "C", "B")

	got, err := composition.SortOperations([]op.Operation{c, a, b})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "pkg.A", got[0].ID)
	require.Equal(t, "pkg.B", got[1].ID)
	require.Equal(t, "pkg.C", got[2].ID)
}

func TestSort_Diamond(t *testing.T) {
	a := opStep("pkg.A", "A")
	b := opStep("pkg.B", "B", "A")
	c := opStep("pkg.C", "C", "A")
	d := opStep("pkg.D", "D", "B", "C")

	got, err := composition.SortOperations([]op.Operation{c, d, a, b})
	require.NoError(t, err)
	require.Len(t, got, 4)
	require.Equal(t, "pkg.A", got[0].ID)
	require.Contains(t, []string{"pkg.B", "pkg.C"}, got[1].ID)
	require.Contains(t, []string{"pkg.B", "pkg.C"}, got[2].ID)
	require.Equal(t, "pkg.D", got[3].ID)
}

func TestSort_Cycle(t *testing.T) {
	a := opStep("pkg.A", "A", "B")
	b := opStep("pkg.B", "B", "A")

	_, err := composition.SortOperations([]op.Operation{a, b})
	require.ErrorIs(t, err, composition.ErrCyclicDependency)
}

func TestSort_SelfDep(t *testing.T) {
	a := opStep("pkg.A", "A", "A")

	_, err := composition.SortOperations([]op.Operation{a})
	require.ErrorIs(t, err, composition.ErrCyclicDependency)
}

func TestSort_MissingDep(t *testing.T) {
	a := opStep("pkg.A", "A", "B")

	got, err := composition.SortOperations([]op.Operation{a})
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "pkg.A", got[0].ID)
}

func TestSort_OrderTrait(t *testing.T) {
	a := opStep("pkg.A", "A")
	b := opStep("pkg.B", "B", "A")
	c := opStep("pkg.C", "C", "B")

	got, err := composition.SortOperations([]op.Operation{c, a, b})
	require.NoError(t, err)
	require.Len(t, got, 3)

	for i, operation := range got {
		found := false

		for _, tr := range operation.Trait {
			if tr.ID == trait.OrderID {
				found = true
				v, ok := tr.Value.(int)
				require.True(t, ok)
				require.Equal(t, i, v)

				break
			}
		}

		require.True(t, found, "step %d missing order trait", i)
	}
}

func TestSort_ImplementsReady(t *testing.T) {
	iface := opWithImplements("pkg.NewLogger", "*pkg.MyLogger", []string{"pkg.Logger"})
	cons := opStep("pkg.NewApp", "*pkg.App", "pkg.Logger")

	got, err := composition.SortOperations([]op.Operation{iface, cons})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "pkg.NewLogger", got[0].ID)
	require.Equal(t, "pkg.NewApp", got[1].ID)
}

func TestSort_Empty(t *testing.T) {
	got, err := composition.SortOperations(nil)
	require.NoError(t, err)
	require.Empty(t, got)
}
