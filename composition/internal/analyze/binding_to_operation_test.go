package analyze_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/composition/data"
	"github.com/thumbrise/gcce/composition/internal/analyze"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
)

type (
	Config   struct{}
	Server   struct{}
	Logger   interface{ Log() }
	MyLogger struct{}
)

func (m *MyLogger) Log() {}

func NewConfig() Config              { return Config{} }
func NewServer(cfg Config) *Server   { return &Server{} }
func NewErrConfig() (*Config, error) { return &Config{}, nil }
func NewMyLogger() *MyLogger         { return &MyLogger{} }

func TestBindingToOperation_Simple(t *testing.T) {
	op, err := analyze.BindingToOperation(&data.Binding{
		Constructor: NewConfig,
	})
	require.NoError(t, err)
	require.NotEmpty(t, op.ID)
	require.NotEmpty(t, op.Trait)
	require.Len(t, op.Output, 1)
	require.NotEmpty(t, op.Output[0].ID)
	require.Empty(t, op.Error)
}

func TestBindingToOperation_DepsInInput(t *testing.T) {
	op, err := analyze.BindingToOperation(&data.Binding{
		Constructor: NewServer,
	})
	require.NoError(t, err)
	require.Len(t, op.Input, 1)
	require.Contains(t, op.Input[0].ID, "Config")
	require.Len(t, op.Output, 1)
}

func TestBindingToOperation_Implements(t *testing.T) {
	op, err := analyze.BindingToOperation(&data.Binding{
		Constructor: NewMyLogger,
		Implements:  []interface{}{(*Logger)(nil)},
	})
	require.NoError(t, err)
	require.Len(t, op.Output, 1)

	found := false

	for _, t := range op.Output[0].Trait {
		if t.ID == trait.ImplementsID {
			found = true

			break
		}
	}

	require.True(t, found)
}

func TestBindingToOperation_WithError(t *testing.T) {
	op, err := analyze.BindingToOperation(&data.Binding{
		Constructor: NewErrConfig,
	})
	require.NoError(t, err)
	require.Len(t, op.Error, 1)
	require.Equal(t, "error", op.Error[0].ID)
}

func TestBindingToOperation_NoError(t *testing.T) {
	op, err := analyze.BindingToOperation(&data.Binding{
		Constructor: NewConfig,
	})
	require.NoError(t, err)
	require.Empty(t, op.Error)
}
