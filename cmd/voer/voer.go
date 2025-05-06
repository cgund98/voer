package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgund98/voer/internal/proto"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a directory path")
	}

	dirPath := os.Args[1]
	var protoFiles []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}

	if len(protoFiles) == 0 {
		log.Fatal("No .proto files found in the specified directory")
	}

	ctx := context.Background()
	files, err := proto.ParsePath(ctx, protoFiles...)
	if err != nil {
		log.Fatalf("Error parsing proto files: %v", err)
	}

	fmt.Printf("Parsed %d files\n", len(files))
}
