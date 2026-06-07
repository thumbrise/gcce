package util_test

import (
	"testing"

	"github.com/thumbrise/gcce/op/emit/pipeline/util"
)

// testTargetFunc is a sample function used by the tests.
func testTargetFunc() {}

func TestFuncDecl_Success(t *testing.T) {
	fd, err := util.FuncDecl(testTargetFunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fd.Name.Name != "testTargetFunc" {
		t.Errorf("expected function name 'testTargetFunc', got '%s'", fd.Name.Name)
	}
}

func TestFuncDecl_NotAFunction(t *testing.T) {
	_, err := util.FuncDecl(42)
	if err == nil {
		t.Fatal("expected error for non-function")
	}
}
