package evidence

import (
	"path"
	"strings"
)

func (c *collector) collectStructure(existingTargets []string) {
	files := sortedFileList(c.files)
	for _, target := range sortedKeys(normalizedSet(existingTargets)) {
		targetDir := path.Dir(target)
		for _, candidate := range files {
			if candidate != target && path.Dir(candidate) == targetDir {
				c.add(candidate, KindSiblingTarget, target, "", 1)
			}
			if testCounterparts(target, candidate) {
				c.add(candidate, KindTestCounterpart, target, "", 1)
			}
		}
	}
}

func (c *collector) collectDependencies(existingTargets []string, edges []DependencyEdge) {
	seeds := normalizedSet(existingTargets)
	for _, edge := range edges {
		source := normalizePath(edge.Source)
		target := normalizePath(edge.Target)
		if _, ok := seeds[source]; ok {
			c.add(target, KindDependencyNeighbor, source, "outbound:"+edge.Relation, 1)
		}
		if _, ok := seeds[target]; ok {
			c.add(source, KindDependencyNeighbor, target, "inbound:"+edge.Relation, 1)
		}
	}
}

func (c *collector) collectRelatedDocuments(documents []RelatedDocument) {
	for _, document := range documents {
		source := normalizePath(document.Path)
		for _, target := range document.Targets {
			c.add(target, KindRelatedDocumentTarget, source, "", 1)
		}
	}
}

func testCounterparts(left, right string) bool {
	leftSubject, leftTest, leftExt := testIdentity(left)
	rightSubject, rightTest, rightExt := testIdentity(right)
	return leftTest != rightTest && leftExt == rightExt && leftSubject != "" && leftSubject == rightSubject
}

func testIdentity(value string) (subject string, test bool, extension string) {
	value = normalizePath(value)
	base := path.Base(value)
	extension = strings.ToLower(path.Ext(base))
	name := strings.TrimSuffix(base, path.Ext(base))
	lower := strings.ToLower(name)

	for _, segment := range strings.Split(strings.ToLower(path.Dir(value)), "/") {
		if segment == "test" || segment == "tests" || segment == "spec" || segment == "specs" {
			test = true
			break
		}
	}
	for _, prefix := range []string{"test_", "spec_"} {
		if strings.HasPrefix(lower, prefix) {
			lower = strings.TrimPrefix(lower, prefix)
			test = true
		}
	}
	for _, suffix := range []string{"_test", "_spec", ".test", ".spec"} {
		if strings.HasSuffix(lower, suffix) {
			lower = strings.TrimSuffix(lower, suffix)
			test = true
		}
	}
	return lower, test, extension
}
