package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/cgund98/voer/internal/proto"
	"github.com/urfave/cli/v3"
)

const (
	// Flag names
	prevFlag   = "previous"
	latestFlag = "latest"
)

// validateAction is the action for the validate command
func validateAction(ctx context.Context, cmd *cli.Command) error {

	prevPath := cmd.String(prevFlag)
	latestPath := cmd.String(latestFlag)

	if prevPath == "" || latestPath == "" {
		return errors.New("previous and latest paths are required")
	}

	prevFiles, err := proto.ParsePath(ctx, prevPath)
	if err != nil {
		return fmt.Errorf("error parsing previous proto files: %v", err)
	}
	if len(prevFiles) != 1 {
		return fmt.Errorf("expected 1 previous proto file, got %d", len(prevFiles))
	}
	prevFile := prevFiles[0]

	latestFiles, err := proto.ParsePath(ctx, latestPath)
	if err != nil {
		return fmt.Errorf("error parsing latest proto files: %v", err)
	}
	if len(latestFiles) != 1 {
		return fmt.Errorf("expected 1 latest proto file, got %d", len(latestFiles))
	}
	latestFile := latestFiles[0]

	prevMessages := proto.ParseMessagesFromFile(prevFile)
	latestMessages := proto.ParseMessagesFromFile(latestFile)

	err = proto.ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err != nil {
		return fmt.Errorf("error validating backwards compatible messages: %v", err)
	}

	fmt.Println("Backwards compatible messages validated successfully")

	return nil
}

// Validate will validate that a proto file is backwards compatible with another
func ValidateCommand() *cli.Command {
	return &cli.Command{
		Name:   "validate",
		Usage:  "Validate that a proto file is backwards compatible with another",
		Action: validateAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     prevFlag,
				Usage:    "The previous proto file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     latestFlag,
				Usage:    "The latest proto file",
				Required: true,
			},
		},
	}
}
