package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseProjectArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arg      string
		wantDir  string
		wantName string
		wantErr  bool
	}{
		{arg: "goapp", wantDir: "goapp", wantName: "goapp"},
		{arg: "codebase/goapp/", wantDir: filepath.Join("codebase", "goapp"), wantName: "goapp"},
		{arg: "/abs/path/my_app-2", wantDir: "/abs/path/my_app-2", wantName: "my_app-2"},
		{arg: "", wantErr: true},
		{arg: ".", wantErr: true},
		{arg: "~", wantErr: true},
		{arg: "2048", wantErr: true},
		{arg: "codebase/my.app", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			t.Parallel()

			dir, name, err := parseProjectArg(tt.arg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseProjectArg(%q) = (%q, %q), want error", tt.arg, dir, name)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseProjectArg(%q) returned error: %v", tt.arg, err)
			}
			if dir != tt.wantDir || name != tt.wantName {
				t.Fatalf("parseProjectArg(%q) = (%q, %q), want (%q, %q)", tt.arg, dir, name, tt.wantDir, tt.wantName)
			}
		})
	}
}

func TestApplyHostPostgresPort(t *testing.T) {
	t.Parallel()

	const compose = "    ports:\n      - 5432:5432\n    environment:\n      POSTGRES_PORT: 5432\n"
	const makefileNoPort = `goose postgres "host=localhost user=postgres password=postgres" up`
	const makefileWithPort = `goose postgres "host=localhost port=5432 user=postgres password=postgres" up`
	const wantCompose = "    ports:\n      - 5544:5432\n    environment:\n      POSTGRES_PORT: 5432\n"
	const wantMakefile = `goose postgres "host=localhost port=5544 user=postgres password=postgres" up`

	t.Run("default port changes nothing", func(t *testing.T) {
		t.Parallel()

		gotCompose, gotMakefile, err := applyHostPostgresPort(compose, makefileNoPort, defaultPostgresHostPort)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotCompose != compose || gotMakefile != makefileNoPort {
			t.Fatalf("content changed for the default port:\n%s\n%s", gotCompose, gotMakefile)
		}
	})

	for name, makefile := range map[string]string{"dsn without port": makefileNoPort, "dsn with port": makefileWithPort} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotCompose, gotMakefile, err := applyHostPostgresPort(compose, makefile, 5544)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotCompose != wantCompose {
				t.Fatalf("compose = %q, want %q", gotCompose, wantCompose)
			}
			if gotMakefile != wantMakefile {
				t.Fatalf("makefile = %q, want %q", gotMakefile, wantMakefile)
			}
		})
	}

	t.Run("missing port mapping is an error", func(t *testing.T) {
		t.Parallel()

		_, _, err := applyHostPostgresPort("services: {}\n", makefileNoPort, 5544)
		if err == nil || !strings.Contains(err.Error(), "5432:5432") {
			t.Fatalf("err = %v, want missing port mapping error", err)
		}
	})
}

func TestRollbackProjectRemovesCreatedParents(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	projectDir := filepath.Join(base, "new", "nested", "goapp")

	rollbackRoot := topmostMissingDir(projectDir)
	if rollbackRoot != filepath.Join(base, "new") {
		t.Fatalf("topmostMissingDir = %q, want %q", rollbackRoot, filepath.Join(base, "new"))
	}

	err := os.MkdirAll(projectDir, 0o755)
	if err != nil {
		t.Fatalf("creating project dir: %v", err)
	}
	err = rollbackProject(t.Context(), projectDir, rollbackRoot, false)
	if err != nil {
		t.Fatalf("rollbackProject returned error: %v", err)
	}
	_, err = os.Stat(rollbackRoot)
	if !os.IsNotExist(err) {
		t.Fatalf("rollback root still exists (stat err = %v)", err)
	}
	_, err = os.Stat(base)
	if err != nil {
		t.Fatalf("pre-existing parent was removed: %v", err)
	}
}
