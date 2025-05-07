package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/cgund98/voer/internal/service/command"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			command.ValidateCommand(),
			command.UploadCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
