package render_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoldenFiles(t *testing.T) {
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not installed, skipping golden tests")
	}

	repoRoot := findRepoRoot(t)
	pluginPath := buildPlugin(t, repoRoot)
	kratosProto := findKratosErrorsProto(t)

	testdata := filepath.Join(repoRoot, "internal", "render", "testdata")

	cases := []struct {
		proto  string
		golden string
	}{
		{"helloworld_error.proto", "helloworld_error.golden.md"},
		{"blog_error.proto", "blog_error.golden.md"},
	}

	for _, tc := range cases {
		t.Run(tc.proto, func(t *testing.T) {
			outDir := t.TempDir()
			cmd := exec.Command("protoc",
				"--plugin=protoc-gen-errors-md="+pluginPath,
				"--errors-md_out="+outDir,
				"--errors-md_opt=paths=source_relative",
				"--proto_path="+testdata,
				"--proto_path="+kratosProto,
				tc.proto,
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("protoc failed: %v\n%s", err, out)
			}
			got, err := os.ReadFile(filepath.Join(outDir, strings.TrimSuffix(tc.proto, ".proto")+".md"))
			if err != nil {
				t.Fatalf("read generated: %v", err)
			}
			goldenPath := filepath.Join(testdata, tc.golden)
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatalf("update golden: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(got) != string(want) {
				t.Fatalf("golden mismatch for %s\n--- want ---\n%s\n--- got ---\n%s", tc.proto, want, got)
			}
		})
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func buildPlugin(t *testing.T, repoRoot string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "protoc-gen-errors-md")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/protoc-gen-errors-md")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build plugin: %v\n%s", err, out)
	}
	return bin
}

func findKratosErrorsProto(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatalf("go env GOMODCACHE: %v", err)
	}
	cache := strings.TrimSpace(string(out))
	matches, err := filepath.Glob(filepath.Join(cache, "github.com", "go-kratos", "kratos", "v2@*", "errors", "errors.proto"))
	if err != nil || len(matches) == 0 {
		t.Skipf("kratos errors.proto not found in module cache")
	}
	return filepath.Dir(filepath.Dir(matches[0]))
}
