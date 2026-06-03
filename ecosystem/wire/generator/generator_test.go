package generator //nolint:testpackage

import (
	"bytes"
	"context"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func opErr(id, target string, deps ...string) op.Operation {
	in := make([]op.Term, len(deps))
	for i, d := range deps {
		in[i] = op.Term{ID: d}
	}

	return op.Operation{
		ID:     id,
		Output: []op.Term{{ID: target}},
		Input:  in,
		Error:  []op.Term{{ID: "error"}},
	}
}

func opIface(id, target string, implements []string, deps ...string) op.Operation {
	implTraits := make([]op.Term, len(implements))
	for i, impl := range implements {
		implTraits[i] = op.Term{
			ID:    trait.ImplementsID,
			Value: impl,
		}
	}

	in := make([]op.Term, len(deps))
	for i, d := range deps {
		in[i] = op.Term{ID: d}
	}

	return op.Operation{
		ID:     id,
		Output: []op.Term{{ID: target, Trait: implTraits}},
		Input:  in,
	}
}

func opVariadic(id, target, dep string) op.Operation {
	return op.Operation{
		ID:     id,
		Output: []op.Term{{ID: target}},
		Input: []op.Term{{
			ID:    dep,
			Trait: []op.Term{trait.NewVariadic()},
		}},
	}
}

func requireValidGo(t *testing.T, src string) {
	t.Helper()

	_, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Fatalf("invalid Go source:\n%s\n\nerror: %v", src, err)
	}
}

func requireContains(t *testing.T, src, substr string) {
	t.Helper()

	if !strings.Contains(src, substr) {
		t.Fatalf("expected %q to contain %q", src, substr)
	}
}

// --- Resolve tests ---

func TestResolve_LinearChain(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewA", "A"),
		opStep("pkg.NewB", "B", "A"),
		opStep("pkg.NewC", "C", "B"),
	}

	factories, err := resolve(ops, "C")
	require.NoError(t, err)
	require.Len(t, factories, 3)
	require.Equal(t, "a", factories[0].varName)
	require.Equal(t, "b", factories[1].varName)
	require.Equal(t, "c", factories[2].varName)
	require.True(t, factories[2].root)
	require.Equal(t, []string{"a"}, factories[1].ctorArgs)
	require.Equal(t, []string{"b"}, factories[2].ctorArgs)
}

func TestResolve_Diamond(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewA", "A"),
		opStep("pkg.NewB", "B", "A"),
		opStep("pkg.NewC", "C", "A"),
		opStep("pkg.NewD", "D", "B", "C"),
	}

	factories, err := resolve(ops, "D")
	require.NoError(t, err)
	require.Len(t, factories, 4)
	require.Equal(t, "d", factories[3].varName)
	require.True(t, factories[3].root)
}

func TestResolve_ViaImplements(t *testing.T) {
	ops := []op.Operation{
		opIface("pkg.NewLogger", "*pkg.MyLogger", []string{"pkg.Logger"}),
		opStep("pkg.NewApp", "*pkg.App", "pkg.Logger"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 2)
	require.Equal(t, "myLogger", factories[0].varName)
	require.Equal(t, "app", factories[1].varName)
	require.Equal(t, []string{"myLogger"}, factories[1].ctorArgs)
}

func TestResolve_RootNotFound(t *testing.T) {
	_, err := resolve([]op.Operation{opStep("pkg.NewA", "A")}, "B")
	require.ErrorIs(t, err, ErrRootNotFound)
}

func TestResolve_EmptyOps(t *testing.T) {
	_, err := resolve(nil, "A")
	require.ErrorIs(t, err, ErrRootNotFound)
}

func TestResolve_UnresolvableDep(t *testing.T) {
	factories, err := resolve([]op.Operation{opStep("pkg.NewA", "A", "B")}, "A")
	require.NoError(t, err)
	require.Len(t, factories, 2)
	require.Equal(t, "pkg.NewA", factories[0].ctor)
	require.True(t, factories[1].external)
	require.Equal(t, "B", factories[1].target)
}

func TestResolve_MultipleProviders(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewOne", "A"),
		opStep("pkg.NewTwo", "A"),
		opStep("pkg.NewApp", "*pkg.App", "A"),
	}

	_, err := resolve(ops, "*pkg.App")
	require.ErrorIs(t, err, ErrMultipleProviders)
}

func TestResolve_IgnoresUnreachable(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewA", "A"),
		opStep("pkg.NewB", "B"),
		opStep("pkg.NewC", "C", "A"),
	}

	factories, err := resolve(ops, "C")
	require.NoError(t, err)
	require.Len(t, factories, 2)
	require.Equal(t, "a", factories[0].varName)
	require.Equal(t, "c", factories[1].varName)
}

func TestResolve_ErrorPropagation(t *testing.T) {
	ops := []op.Operation{
		opErr("pkg.NewA", "A"),
		opStep("pkg.NewApp", "*pkg.App", "A"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 2)
	require.True(t, factories[0].returnsErr)
	require.False(t, factories[1].returnsErr)
}

func TestResolve_VarNameCollision(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewA", "pkg.A"),
		opStep("pkg2.NewA", "pkg2.A"),
		opStep("pkg.NewApp", "*pkg.App", "pkg.A", "pkg2.A"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 3)
	require.Equal(t, "a", factories[0].varName)
	require.Equal(t, "a1", factories[1].varName)
}

func TestResolve_ImportAliasCollision(t *testing.T) {
	ops := []op.Operation{
		opStep("custom/http.NewServer", "*custom/http.Server"),
		opStep("default/http.NewClient", "*default/http.Client"),
		opStep("pkg.NewApp", "*pkg.App", "*custom/http.Server", "*default/http.Client"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 3)

	// aliases: http (custom/http), http1 (default/http)
	// varName from Server → server, Client → client — no collision with http
	require.Equal(t, "server", factories[0].varName)
	require.Equal(t, "client", factories[1].varName)
}

func TestResolve_GoKeywordVarName(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewType", "pkg.Type"),
		opStep("pkg.NewApp", "*pkg.App", "pkg.Type"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 2)
	// varName from Type → "type", but "type" is go keyword
	require.Equal(t, "type1", factories[0].varName)
	require.Equal(t, "app", factories[1].varName)
}

func TestResolve_ErrVarNameReserved(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewErr", "pkg.Err"),
		opStep("pkg.NewApp", "*pkg.App", "pkg.Err"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	// varName from Err → "err", but "err" is in used set
	require.Equal(t, "err1", factories[0].varName)
	require.Equal(t, "app", factories[1].varName)
}

func TestResolve_VariadicDep(t *testing.T) {
	ops := []op.Operation{
		opVariadic("pkg.NewClient", "*pkg.Client", "[]pkg.Plugin"),
		opStep("pkg.NewApp", "*pkg.App", "*pkg.Client"),
	}

	factories, err := resolve(ops, "*pkg.App")
	require.NoError(t, err)
	require.Len(t, factories, 3)
	require.True(t, factories[0].variadic)
	require.Equal(t, "client", factories[0].varName)
	require.Len(t, factories[0].ctorArgs, 1)
	require.Equal(t, "plugin", factories[0].ctorArgs[0])
}

// --- Codegen tests ---

func TestCodegen_Simple(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "cfg", ctor: "pkg.NewConfig"},
		{varName: "app", ctor: "pkg.NewApp", ctorArgs: []string{"cfg"}, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.App"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func Provide() *pkg.App")
	requireContains(t, buf.String(), "return app")
}

func TestCodegen_WithError(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "cfg", ctor: "pkg.NewConfig", returnsErr: true, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.Config"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func Provide() (*pkg.Config, error)")
	requireContains(t, buf.String(), "if err != nil")
	requireContains(t, buf.String(), "return cfg, nil")
}

func TestCodegen_Variadic(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "plugins", ctor: "pkg.NewPlugins"},
		{varName: "client", ctor: "pkg.NewClient", ctorArgs: []string{"plugins"}, variadic: true},
		{varName: "app", ctor: "pkg.NewApp", ctorArgs: []string{"client"}, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.App"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "client := pkg.NewClient(plugins...)")
}

func TestCodegen_MultipleSteps(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "cfg", ctor: "pkg.NewConfig"},
		{varName: "srv", ctor: "pkg.NewServer", ctorArgs: []string{"cfg"}},
		{varName: "app", ctor: "pkg.NewApp", ctorArgs: []string{"srv"}, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.App"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "cfg := pkg.NewConfig()")
	requireContains(t, buf.String(), "app := pkg.NewApp(srv)")
	requireContains(t, buf.String(), "return app")
}

func TestCodegen_ErrorMid(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "cfg", ctor: "pkg.NewConfig", returnsErr: true},
		{varName: "srv", ctor: "pkg.NewServer", ctorArgs: []string{"cfg"}, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.Server"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "cfg, err := pkg.NewConfig()")
	requireContains(t, buf.String(), "if err != nil {")
	requireContains(t, buf.String(), "return nil, err")
	requireContains(t, buf.String(), "srv := pkg.NewServer(cfg)")
	requireContains(t, buf.String(), "return srv, nil")
}

func TestCodegen_CrossPackage(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "mux", ctor: "net/http.NewServeMux"},
		{varName: "app", ctor: "pkg.NewApp", ctorArgs: []string{"mux"}, root: true},
	}, Config{PkgName: "main", FuncName: "Provide", RootType: "*pkg.App"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())

	src := buf.String()
	requireContains(t, src, "net/http")
	requireContains(t, src, "mux := http.NewServeMux()")
	requireContains(t, src, "app := pkg.NewApp(mux)")
}

func TestCodegen_CustomPackage(t *testing.T) {
	var buf bytes.Buffer

	err := generateCode(&buf, []factory{
		{varName: "app", ctor: "pkg.NewApp", root: true},
	}, Config{PkgName: "custom", FuncName: "Init", RootType: "*pkg.App"})

	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "package custom")
	requireContains(t, buf.String(), "func Init() *pkg.App")
}

// --- Integration tests ---

func TestGen_SimpleChain(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewConfig", "Config"),
		opStep("pkg.NewServer", "*pkg.Server", "Config"),
		opStep("pkg.NewHandler", "*pkg.Handler", "*pkg.Server"),
		opStep("pkg.NewApp", "*pkg.App", "*pkg.Handler"),
	}

	var buf bytes.Buffer

	err := Generate(&buf, ops, Config{
		PkgName:  "main",
		FuncName: "ProvideApp",
		RootType: "*pkg.App",
	})
	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func ProvideApp() *pkg.App")
	requireContains(t, buf.String(), "return app")
}

func TestGen_ErrorChain(t *testing.T) {
	ops := []op.Operation{
		opStep("pkg.NewConfig", "Config"),
		opErr("pkg.NewServer", "*pkg.Server", "Config"),
		opStep("pkg.NewApp", "*pkg.App", "*pkg.Server"),
	}

	var buf bytes.Buffer

	err := Generate(&buf, ops, Config{
		PkgName:  "main",
		FuncName: "ProvideApp",
		RootType: "*pkg.App",
	})
	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "error")
	requireContains(t, buf.String(), "if err != nil")
}

func TestGen_WithImplements(t *testing.T) {
	ops := []op.Operation{
		opIface("pkg.NewLogger", "*pkg.MyLogger", []string{"pkg.Logger"}),
		opStep("pkg.NewApp", "*pkg.App", "pkg.Logger"),
	}

	var buf bytes.Buffer

	err := Generate(&buf, ops, Config{
		PkgName:  "main",
		FuncName: "ProvideApp",
		RootType: "*pkg.App",
	})
	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func ProvideApp() *pkg.App")
}

func TestGen_RootNotFound(t *testing.T) {
	err := Generate(nil, []op.Operation{opStep("pkg.NewA", "A")}, Config{
		PkgName:  "main",
		FuncName: "Provide",
		RootType: "B",
	})
	require.ErrorIs(t, err, ErrRootNotFound)
}

func TestGen_ExternalDep(t *testing.T) {
	var buf bytes.Buffer

	err := Generate(&buf, []op.Operation{opStep("pkg.NewA", "A", "B")}, Config{
		PkgName:  "main",
		FuncName: "Provide",
		RootType: "A",
	})
	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func Provide(b B) A")
	requireContains(t, buf.String(), "return a")
}

func TestGen_VariadicDep(t *testing.T) {
	var buf bytes.Buffer

	err := Generate(&buf, []op.Operation{
		opVariadic("pkg.NewClient", "*pkg.Client", "[]pkg.Plugin"),
		opStep("pkg.NewApp", "*pkg.App", "*pkg.Client"),
	}, Config{
		PkgName:  "main",
		FuncName: "ProvideApp",
		RootType: "*pkg.App",
	})
	require.NoError(t, err)
	requireValidGo(t, buf.String())
	requireContains(t, buf.String(), "func ProvideApp(plugin []pkg.Plugin) *pkg.App")
	requireContains(t, buf.String(), "client := pkg.NewClient(plugin...)")
}

func TestCodegen_ErrorWithValueRoot_BuildFails(t *testing.T) {
	factories := []factory{
		{varName: "cfg", ctor: "NewConfig", returnsErr: true, root: true},
	}

	dir := t.TempDir()

	genPath := filepath.Join(dir, "wire_gen.go")
	genFile, err := os.Create(genPath)
	require.NoError(t, err)

	err = generateCode(genFile, factories, Config{
		PkgName:  "main",
		FuncName: "Provide",
		RootType: "Config",
	})
	require.NoError(t, err)
	require.NoError(t, genFile.Close())

	err = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module wiretest\ngo 1.22\n"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "pkg.go"), []byte("package main\ntype Config struct{}\nfunc NewConfig() (Config, error) { return Config{}, nil }\nfunc main() { Provide() }\n"), 0o644)
	require.NoError(t, err)

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "out"))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err, "generated code with value root + error should compile: %s", string(out))
}
