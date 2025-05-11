package command

import (
	"context"
	"fmt"
	"net"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/infra/logging"
	"github.com/cgund98/voer/internal/infra/sqlite"
	"github.com/cgund98/voer/internal/service/frontend"
	svc "github.com/cgund98/voer/internal/service/grpc"
)

const (
	// Flag names
	grpcPortFlag     = "grpc-port"
	frontendPortFlag = "frontend-port"
)

// serverAction is the action for the port command
func serverAction(ctx context.Context, config *config.Config, cmd *cli.Command) error {
	// Flags
	grpcPort := cmd.Int(grpcPortFlag)
	frontendPort := cmd.Int(frontendPortFlag)

	// Initialize DB connection
	db, err := sqlite.NewDB(config.SqliteDBPath)
	if err != nil {
		return fmt.Errorf("error initializing DB connection: %v", err)
	}

	// Initialize gRPC server
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(svc.LoggerInterceptor))

	// Register services
	v1.RegisterPackageSvcServer(grpcServer, &svc.PackageSvc{
		DB: db,
	})

	// Start frontend and gRPC servers in parallel with an ErrGroup
	eg, _ := errgroup.WithContext(ctx)

	// Start frontend service
	frontendSvc := frontend.NewService(config, db)
	frontendSvc.Init()

	eg.Go(func() error {
		return frontendSvc.Start(frontendPort)
	})

	// Start gRPC server
	eg.Go(func() error {
		logging.Logger.Info("Starting gRPC server...", "port", grpcPort)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			return fmt.Errorf("failed to listen: %v", err)
		}

		// Start server
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("failed to serve: %v", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error encountered by server: %v", err)
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
				Name:     grpcPortFlag,
				Usage:    "The port to start the gRPC service on",
				Required: false,
				Value:    config.GrpcPort,
			},
			&cli.IntFlag{
				Name:     frontendPortFlag,
				Usage:    "The port to start the frontend service on",
				Required: false,
				Value:    config.FrontendPort,
			},
		},
	}
}
