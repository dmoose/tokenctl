package build

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBuildCSS_Golden(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"valid", "extends"} {
		res, err := BuildCSS(dirFS(t, "../../testdata/fixtures/"+name), Options{Format: FormatTailwind})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		golden, err := os.ReadFile("../../testdata/golden/" + name + ".css")
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(res.CSS) != strings.TrimSpace(string(golden)) {
			t.Errorf("%s: output differs from golden", name)
		}
	}
}

// BuildCSS over os.DirFS and BuildCSSDirs must agree on every tree in the repo.
func TestBuildCSS_FSMatchesDirs(t *testing.T) {
	t.Parallel()
	dirs, _ := filepath.Glob("../../testdata/fixtures/*")
	examples, _ := filepath.Glob("../../examples/*")
	for _, dir := range append(dirs, examples...) {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		for _, f := range []Format{FormatCSS, FormatTailwind} {
			want, wantErr := BuildCSSDirs([]string{dir}, Options{Format: f})
			got, gotErr := BuildCSS(dirFS(t, dir), Options{Format: f})
			if (wantErr == nil) != (gotErr == nil) {
				t.Errorf("%s: dirs err %v, fs err %v", dir, wantErr, gotErr)
				continue
			}
			if got.CSS != want.CSS {
				t.Errorf("%s format %d: fs output differs from dirs output", dir, f)
			}
		}
	}
}

func TestBuildCSS_MergeReportsConflicts(t *testing.T) {
	t.Parallel()
	trees := append(dirFS(t, "../../testdata/fixtures/merge-base"), dirFS(t, "../../testdata/fixtures/merge-ext")...)
	res, err := BuildCSS(trees, Options{})
	if err != nil {
		t.Fatal(err)
	}
	want, err := BuildCSSDirs([]string{"../../testdata/fixtures/merge-base", "../../testdata/fixtures/merge-ext"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.CSS != want.CSS {
		t.Error("merged fs output differs from merged dirs output")
	}
	if len(res.Conflicts) == 0 {
		t.Error("expected conflicts from merge-ext overriding merge-base")
	}
}

func TestBuildCSS_MapFS(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"brand/colors.json": {Data: []byte(`{"color": {"blue": {"$value": "#0000ff"}}}`)},
		"themes/dark.json":  {Data: []byte(`{"dark": {"color": {"blue": {"$value": "#000080"}}}}`)},
	}
	res, err := BuildCSS([]fs.FS{fsys}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.CSS, "#0000ff") || !strings.Contains(res.CSS, "#000080") {
		t.Errorf("expected base and dark theme values in output:\n%s", res.CSS)
	}
}

func TestBuildCSS_StrictUnknownKeys(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"components/btn.json": {Data: []byte(`{"button": {"$class": "btn", "$base": {"color": "red"}, "variant": {}}}`)},
	}
	res, err := BuildCSS([]fs.FS{fsys}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) == 0 {
		t.Fatal("expected a finding for the unknown component key")
	}
	_, err = BuildCSS([]fs.FS{fsys}, Options{StrictUnknownKeys: true})
	if !errors.Is(err, ErrUnknownKeys) {
		t.Errorf("strict build: got %v, want ErrUnknownKeys", err)
	}
}

func dirFS(t *testing.T, dir string) []fs.FS {
	t.Helper()
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
	return []fs.FS{os.DirFS(dir)}
}
