package proto

import "github.com/bufbuild/protocompile/linker"

// GroupByPackage groups the files by package name
func GroupByPackage(files linker.Files) map[string]linker.Files {
	packages := make(map[string]linker.Files)
	for _, file := range files {
		packages[string(file.Package())] = append(packages[string(file.Package())], file)
	}
	return packages
}
