package struc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thumbrise/gcce/op/emit/struc"
	"github.com/thumbrise/gcce/pkg/op-composition-go/trait"
)

type AuthService struct{}

func (AuthService) Login() string { return "ok" }

//nolint:unused
func (AuthService) logout()           {}
func (AuthService) QueryMe(id string) {}

type EmptyService struct{}

func TestT_AuthService(t *testing.T) {
	ops, err := struc.T(AuthService{})
	require.NoError(t, err)
	require.Len(t, ops, 2)

	expectedIDs := []string{
		"github.com/thumbrise/gcce/op/emit/struc_test.(*AuthService).Login",
		"github.com/thumbrise/gcce/op/emit/struc_test.(*AuthService).QueryMe",
	}

	actualIDs := make([]string, 0, len(expectedIDs))
	for _, op := range ops {
		actualIDs = append(actualIDs, op.ID)
		require.Len(t, op.Trait, 1)
		assert.Equal(t, trait.NewGroup("AuthService"), op.Trait[0])
	}

	assert.ElementsMatch(t, expectedIDs, actualIDs)
}

func TestT_ExportedOnly(t *testing.T) {
	ops, err := struc.T(AuthService{})
	require.NoError(t, err)
	assert.Len(t, ops, 2)
}

func TestT_EmptyStruct(t *testing.T) {
	ops, err := struc.T(EmptyService{})
	require.NoError(t, err)
	assert.Empty(t, ops)
}

func TestT_PointerReceiver(t *testing.T) {
	ops, err := struc.T(&AuthService{})
	require.NoError(t, err)
	assert.Len(t, ops, 2)
}

func TestT_NotStruct(t *testing.T) {
	_, err := struc.T("not a struct")
	require.Error(t, err)
}

func TestT_Nil(t *testing.T) {
	_, err := struc.T(nil)
	require.ErrorIs(t, err, struc.ErrNil)
}

// Add to struc_test.go

type ServiceWithPtr struct{}

func (*ServiceWithPtr) Process() string { return "ok" }

type ValueAndPtrService struct{}

func (v ValueAndPtrService) ByValue() string { return "" }
func (v *ValueAndPtrService) ByPtr() string  { return "" }

//nolint:unused
func (v *ValueAndPtrService) private() {}

func TestT_ValueWithPointerMethod(t *testing.T) {
	ops, err := struc.T(ServiceWithPtr{})
	require.NoError(t, err)
	require.Len(t, ops, 1)
	assert.Contains(t, ops[0].ID, "Process")
	require.Len(t, ops[0].Trait, 1)
	assert.Equal(t, trait.NewGroup("ServiceWithPtr"), ops[0].Trait[0])
}

func TestT_PointerWithPointerMethod(t *testing.T) {
	ops, err := struc.T(&ServiceWithPtr{})
	require.NoError(t, err)
	require.Len(t, ops, 1)
	assert.Contains(t, ops[0].ID, "Process")
}

func TestT_ValueWithBothReceivers(t *testing.T) {
	ops, err := struc.T(ValueAndPtrService{})
	require.NoError(t, err)
	require.Len(t, ops, 2)

	expectedIDs := []string{
		"github.com/thumbrise/gcce/op/emit/struc_test.(*ValueAndPtrService).ByValue",
		"github.com/thumbrise/gcce/op/emit/struc_test.(*ValueAndPtrService).ByPtr",
	}

	actualIDs := make([]string, 0, len(expectedIDs))
	for _, op := range ops {
		actualIDs = append(actualIDs, op.ID)
		require.Len(t, op.Trait, 1)
		assert.Equal(t, trait.NewGroup("ValueAndPtrService"), op.Trait[0])
	}

	assert.ElementsMatch(t, expectedIDs, actualIDs)
}
