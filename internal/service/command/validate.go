package command

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/bufbuild/protocompile/linker"
	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/proto"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// validateAction is the action for the validate command
func validateAction(ctx context.Context, cmd *cli.Command) error {

	protoPath := cmd.String(protoFlag)
	endpoint := cmd.String(endpointFlag)

	if protoPath == "" {
		return errors.New("proto path is required")
	}

	// Init client
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	client := v1.NewPackageSvcClient(conn)

	// Scan for .proto files under the given path
	filePaths, err := findProtoFiles(protoPath)
	if err != nil {
		return err
	}

	protoFiles := make(linker.Files, 0)
	for _, filePath := range filePaths {
		curFiles, err := proto.ParsePath(ctx, filePath)
		if err != nil {
			return fmt.Errorf("error parsing proto files: %v", err)
		}
		protoFiles = append(protoFiles, curFiles...)
	}

	// Validate package names are unique
	err = proto.ValidatePackagesInSameDirectory(ctx, protoFiles)
	if err != nil {
		return err
	}

	// Upload the proto files
	validateReq := &v1.ValidatePackageVersionRequest{}

	// Group based on package name
	packageFiles := proto.GroupByPackage(protoFiles)
	for packageName, files := range packageFiles {

		// Get file contents
		packageFiles := make([]*v1.ProtoFile, 0)
		for _, file := range files {
			fileContents, err := proto.ReadStrings(ctx, file.Path())
			if err != nil {
				return fmt.Errorf("error reading proto files: %v", err)
			}

			packageFiles = append(packageFiles, &v1.ProtoFile{
				FileName:     filepath.Base(file.Path()),
				FileContents: fileContents[0].FileContents,
			})
		}

		validateReq.Packages = append(validateReq.Packages, &v1.PackageFile{
			PackageName: packageName,
			Files:       packageFiles,
		})
	}

	// Validate the proto files
	validateRes, err := client.ValidatePackageVersion(ctx, validateReq)
	if err != nil {
		return fmt.Errorf("error validating proto files: %v", err)
	}

	if validateRes.IsValid {
		fmt.Println("Schema validated successfully")
	} else {
		fmt.Println("Schema is not backwards compatible.")
		fmt.Printf("Error: %v\n", validateRes.Error)
	}

	return nil
}

// Validate will validate that a proto file is backwards compatible with another
func ValidateCommand(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:   "validate",
		Usage:  "Validate that a proto file is backwards compatible with any existing packages",
		Action: validateAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     protoFlag,
				Usage:    "Path to the proto files to validate",
				Required: true,
			},
			&cli.StringFlag{
				Name:     endpointFlag,
				Usage:    "The endpoint to upload the proto file to",
				Required: false,
				Value:    config.GrpcEndpoint,
			},
		},
	}
}
