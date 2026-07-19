package codemapcorpus

import "path"

func repositoryPaths(files []string) []string {
	paths := make(map[string]struct{}, len(files)*2)
	for _, file := range files {
		file = normalizePath(file)
		if file == "" {
			continue
		}
		paths[file] = struct{}{}
		for directory := path.Dir(file); directory != "." && directory != "/"; directory = path.Dir(directory) {
			paths[directory+"/"] = struct{}{}
		}
	}
	return sortedSet(paths)
}
