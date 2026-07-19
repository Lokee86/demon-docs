package codemapcorpus

import (
	"reflect"
	"testing"
)

func TestRepositoryPathsIncludesTrackedParentDirectories(t *testing.T) {
	got := repositoryPaths([]string{
		"services/diagnostic-aggregator/cmd/diagnostic-submit/main.go",
		"services/diagnostic-aggregator/hosted/host.go",
	})
	for _, expected := range []string{
		"services/",
		"services/diagnostic-aggregator/",
		"services/diagnostic-aggregator/cmd/",
		"services/diagnostic-aggregator/cmd/diagnostic-submit/",
		"services/diagnostic-aggregator/cmd/diagnostic-submit/main.go",
	} {
		if !contains(got, expected) {
			t.Fatalf("repository paths omitted %q: %v", expected, got)
		}
	}
	if !reflect.DeepEqual(got, sortedSet(func() map[string]struct{} {
		set := make(map[string]struct{}, len(got))
		for _, value := range got {
			set[value] = struct{}{}
		}
		return set
	}())) {
		t.Fatalf("repository paths are not sorted and unique: %v", got)
	}
}
