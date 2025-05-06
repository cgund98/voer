package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cgund98/voer/internal/proto"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Please provide a previous and latest path")
	}

	prevPath := os.Args[1]
	latestPath := os.Args[2]

	ctx := context.Background()

	prevFiles, err := proto.ParsePath(ctx, prevPath)
	if err != nil {
		log.Fatalf("Error parsing previous proto files: %v", err)
	}
	prevFile := prevFiles[0]

	latestFiles, err := proto.ParsePath(ctx, latestPath)
	if err != nil {
		log.Fatalf("Error parsing latest proto files: %v", err)
	}
	latestFile := latestFiles[0]

	err = proto.ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err != nil {
		log.Fatalf("Error validating backwards compatible messages: %v", err)
	}

	fmt.Println("Backwards compatible messages validated successfully")
}
