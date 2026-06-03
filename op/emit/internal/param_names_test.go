package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit/internal"
)

// Test functions for ParamNames
func NamedParams(a string, b int, c bool)       {}
func UnnamedParams(string, int)                 {}
func MixedParams(name string, age int, _ bool)  {}
func SingleParam(x float64)                     {}
func NoParams()                                 {}
func VariadicParams(prefix string, nums ...int) {}

func TestParamNames_Named(t *testing.T) {
	names, err := internal.ParamNames(NamedParams)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, names)
}

func TestParamNames_Unnamed(t *testing.T) {
	names, err := internal.ParamNames(UnnamedParams)
	require.NoError(t, err)
	assert.Equal(t, []string{"", ""}, names)
}

func TestParamNames_Mixed(t *testing.T) {
	names, err := internal.ParamNames(MixedParams)
	require.NoError(t, err)
	assert.Equal(t, []string{"name", "age", "_"}, names)
}

func TestParamNames_Single(t *testing.T) {
	names, err := internal.ParamNames(SingleParam)
	require.NoError(t, err)
	assert.Equal(t, []string{"x"}, names)
}

func TestParamNames_NoParams(t *testing.T) {
	names, err := internal.ParamNames(NoParams)
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestParamNames_Variadic(t *testing.T) {
	names, err := internal.ParamNames(VariadicParams)
	require.NoError(t, err)
	assert.Equal(t, []string{"prefix", "nums"}, names)
}

func TestParamNames_AnonymousFunction(t *testing.T) {
	fn := func() {}
	_, err := internal.ParamNames(fn)
	require.ErrorIs(t, err, internal.ErrAnonymousFunction)
}

func TestParamNames_NonFunction(t *testing.T) {
	_, err := internal.ParamNames("not a func")
	require.ErrorIs(t, err, internal.ErrNotFunc)
}
