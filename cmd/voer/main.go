package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/service/command"
)

func main() {

	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd := &cli.Command{
		Commands: []*cli.Command{
			command.ValidateCommand(),
			command.UploadCommand(config),
			command.ServerCommand(config),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
