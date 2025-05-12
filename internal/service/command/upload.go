package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bufbuild/protocompile/linker"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/proto"
)

const (
	// Flag names
	protoFlag    = "proto"
	endpointFlag = "endpoint"
)

// findProtoFiles will find all the proto files in a given path
func findProtoFiles(rootPath string) ([]string, error) {
	// Look for all .proto files recursively under the given path
	var protoFiles []string

	// Walk through all files under rootPath
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if file has .proto extension
		if filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking path: %w", err)
	}

	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no proto files found in %s", rootPath)
	}

	return protoFiles, nil
}

// uploadAction is the action for the upload command
func uploadAction(ctx context.Context, cmd *cli.Command) error {
	// Flags
	protoPath := cmd.String(protoFlag)
	endpoint := cmd.String(endpointFlag)

	if protoPath == "" {
		return errors.New("proto file path is required")
	}

	// Scan for .proto files under the given path
	filePaths, err := findProtoFiles(protoPath)
	if err != nil {
		return err
	}

	// Parse the proto files
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

	// Init client
	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		return fmt.Errorf("error creating client: %v", err)
	}
	client := v1.NewPackageSvcClient(conn)

	// Upload the proto files
	uploadReq := &v1.UploadPackageVersionRequest{}

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

		uploadReq.Packages = append(uploadReq.Packages, &v1.PackageFile{
			PackageName: packageName,
			Files:       packageFiles,
		})
	}

	// Upload the proto files
	uploadRes, err := client.UploadPackageVersion(ctx, uploadReq)
	if err != nil {
		return fmt.Errorf("error uploading proto files: %v", err)
	}

	fmt.Println("Uploaded proto files successfully.")
	for _, pkgVer := range uploadRes.PackageVersions {
		fmt.Printf("Created new version of package #%d with version %d\n", pkgVer.PackageId, pkgVer.Version)
	}

	return nil
}

// UploadCommand will upload a set of proto files to the vör service
func UploadCommand(config *config.Config) *cli.Command {
	return &cli.Command{
		Name:   "upload",
		Usage:  "Upload a proto file to the vör service",
		Action: uploadAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     protoFlag,
				Usage:    "The proto file to upload",
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
