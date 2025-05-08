package command

import (
	"context"
	"fmt"
	"net"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"

	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/infra/sqlite"
	svc "github.com/cgund98/voer/internal/service/grpc"
)

const (
	// Flag names
	portFlag = "port"
)

// serverAction is the action for the port command
func serverAction(ctx context.Context, config *config.Config, cmd *cli.Command) error {
	// Flags
	port := cmd.Int(portFlag)

	// Initialize DB connection
	db, err := sqlite.NewDB(config.SqliteDBPath)
	if err != nil {
		return fmt.Errorf("error initializing DB connection: %v", err)
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	v1.RegisterPackageSvcServer(grpcServer, &svc.PackageSvc{
		DB: db,
	})

	// Start server
	fmt.Printf("Starting server on port %d\n", port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Start server
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func makeServerAction(config *config.Config) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return serverAction(ctx, config, cmd)
	}
}

// ServerCommand will start the vör service
func ServerCommand(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:   "server",
		Usage:  "Start the vör service",
		Action: makeServerAction(config),
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     portFlag,
				Usage:    "The port to start the service on",
				Required: false,
				Value:    8000,
			},
		},
	}
}
