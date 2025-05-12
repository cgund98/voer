package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	outputFlag  = "output"
	packageFlag = "package"
	versionFlag = "version"
)

// downloadAction is the action for the download command
func downloadAction(ctx context.Context, cmd *cli.Command) error {

	endpoint := cmd.String(endpointFlag)
	outputDir := cmd.String(outputFlag)
	packageName := cmd.String(packageFlag)
	version := cmd.Uint64(versionFlag)

	if outputDir == "" {
		return errors.New("output directory is required")
	}

	// Validate the output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory does not exist: %v", err)
	}

	if packageName == "" {
		return errors.New("package name is required")
	}

	// Init client
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	client := v1.NewPackageSvcClient(conn)

	// Upload the proto files
	getReq := &v1.GetPackageVersionRequest{
		PackageName: packageName,
		Version:     version,
	}

	// Download the proto files
	downloadRes, err := client.GetPackageVersion(ctx, getReq)
	if err != nil {
		return fmt.Errorf("error validating proto files: %v", err)
	}

	// Write the proto files to the output directory
	for _, file := range downloadRes.Files {
		filePath := filepath.Join(outputDir, file.FileName)

		fmt.Printf("Writing package file '%s' to '%s'\n", file.FileName, filePath)

		err = os.WriteFile(filePath, []byte(file.ProtoContents), 0644)
		if err != nil {
			return fmt.Errorf("error writing proto file: %v", err)
		}
	}

	fmt.Println("Downloaded package files successfully")
	return nil
}

// Download will download that a proto file is backwards compatible with another
func DownloadCommand(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:   "download",
		Usage:  "Download that a proto file is backwards compatible with any existing packages",
		Action: downloadAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     endpointFlag,
				Usage:    "The endpoint to upload the proto file to",
				Required: false,
				Value:    config.GrpcEndpoint,
			},
			&cli.StringFlag{
				Name:     outputFlag,
				Usage:    "The output directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:     packageFlag,
				Usage:    "The package name",
				Required: true,
			},
			&cli.Uint64Flag{
				Name:     versionFlag,
				Usage:    "The version of the package",
				Required: true,
			},
		},
	}
}
