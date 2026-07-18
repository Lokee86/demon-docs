package reconcile

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/demon-docs/internal/config"
)

func TestSameFolderDirectAndStubTransitionsPreserveDescriptions(t *testing.T) {
	tests := []struct {
		name, index, current, want string
		absent                     string
	}{
		{
			name:    "direct_to_stub",
			index:   managedIndex("Docs", "- [foo.md](foo.md) - Custom foo description.\n", "", ""),
			current: filepath.Join("stubs", "foo.md"),
			want:    "- [foo.md](stubs/foo.md) - Stub: Custom foo description.",
			absent:  "foo.md](foo.md) - Custom foo description.",
		},
		{
			name:    "stub_graduation",
			index:   managedIndex("Docs", "", "- [foo.md](stubs/foo.md) - Stub: lower-case promoted description.\n", ""),
			current: "foo.md",
			want:    "- [foo.md](foo.md) - Lower-case promoted description.",
			absent:  "stubs/foo.md",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "docs")
			write(t, filepath.Join(root, "README.md"), test.index)
			write(t, filepath.Join(root, test.current), "# Foo")
			result, err := Tree(root, config.Default())
			if err != nil {
				t.Fatal(err)
			}
			index := plannedText(t, result, filepath.Join(root, "README.md"))
			requireContains(t, index, test.want)
			if strings.Contains(index, test.absent) {
				t.Fatalf("old transition target remained:\n%s", index)
			}
		})
	}
}

func TestCrossFolderFileMoveMatching(t *testing.T) {
	tests := []struct {
		name         string
		stale        map[string]string
		want, absent string
	}{
		{
			name:  "unique_reuses_description",
			stale: map[string]string{"alpha": "Custom alpha description."},
			want:  "Custom alpha description.", absent: "Foo documentation.",
		},
		{
			name:  "ambiguous_uses_fallback",
			stale: map[string]string{"alpha": "Custom alpha description.", "beta": "Custom beta description."},
			want:  "Foo documentation.", absent: "Custom alpha description.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "docs")
			write(t, filepath.Join(root, "README.md"), "# Docs")
			for folder, description := range test.stale {
				write(t, filepath.Join(root, folder, "README.md"), managedIndex(strings.ToUpper(folder[:1])+folder[1:], "- [foo.md](foo.md) - "+description+"\n", "", ""))
			}
			destination := "gamma"
			if len(test.stale) == 1 {
				destination = "beta"
			}
			write(t, filepath.Join(root, destination, "foo.md"), "# Foo")
			result, err := Tree(root, config.Default())
			if err != nil {
				t.Fatal(err)
			}
			index := plannedText(t, result, filepath.Join(root, destination, "README.md"))
			requireContains(t, index, test.want)
			if strings.Contains(index, test.absent) {
				t.Fatalf("unexpected description:\n%s", index)
			}
		})
	}
}

func TestCrossFolderFolderMoveMatching(t *testing.T) {
	tests := []struct {
		name         string
		stale        map[string]string
		want, absent string
	}{
		{"unique_reuses_description", map[string]string{"alpha": "Custom guide description."}, "Custom guide description.", "Guide documentation."},
		{"ambiguous_uses_fallback", map[string]string{"alpha": "Custom alpha guide.", "beta": "Custom beta guide."}, "Guide documentation.", "Custom alpha guide."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "docs")
			write(t, filepath.Join(root, "README.md"), "# Docs")
			for folder, description := range test.stale {
				write(t, filepath.Join(root, folder, "README.md"), managedIndex(strings.ToUpper(folder[:1])+folder[1:], "", "", "- [Guide](guide/README.md) - "+description+"\n"))
			}
			destination := "gamma"
			if len(test.stale) == 1 {
				destination = "beta"
			}
			write(t, filepath.Join(root, destination, "guide", "README.md"), "# Guide")
			result, err := Tree(root, config.Default())
			if err != nil {
				t.Fatal(err)
			}
			index := plannedText(t, result, filepath.Join(root, destination, "README.md"))
			requireContains(t, index, test.want)
			if strings.Contains(index, test.absent) {
				t.Fatalf("unexpected description:\n%s", index)
			}
		})
	}
}
