package proto

import (
	"context"
	"fmt"
	"os"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
)

// ParsePath will look for proto files in under a specific path
func ParsePath(ctx context.Context, filePaths ...string) (linker.Files, error) {

	parser := &protocompile.Compiler{
		// You can add ImportPaths if your .proto imports others
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{}),
	}

	// Compile one or more .proto files
	files, err := parser.Compile(ctx, filePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto file: %w", err)
	}

	return files, nil
}

// ParseString will parse a proto file from a string
func ParseString(ctx context.Context, content string) (linker.File, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "proto")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			fmt.Printf("failed to remove temporary file: %v\n", err)
		}
	}()

	parser := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{}),
	}

	files, err := parser.Compile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto file: %w", err)
	}
	file := files[0]

	return file, nil

}
