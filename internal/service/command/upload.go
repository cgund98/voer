package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bufbuild/protocompile/linker"
	"github.com/urfave/cli/v3"

	ent "github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/infra/sqlite"
	"github.com/cgund98/voer/internal/proto"
)

const (
	// Flag names
	protoFlag = "proto"
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
	err = proto.ValidateUniquePackage(ctx, protoFiles)
	if err != nil {
		return err
	}

	// Init config
	config, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Initialize DB connection
	db, err := sqlite.NewDB(config.SqliteDBPath)
	if err != nil {
		return fmt.Errorf("error initializing DB connection: %v", err)
	}

	// Create transaction
	tx := db.Begin()

	for _, file := range protoFiles {

		// Create a new package
		pkg := ent.Package{PackageName: string(file.Package().Name())}

		// Upsert package
		result := tx.Where(ent.Package{PackageName: pkg.PackageName}).FirstOrCreate(&pkg)
		if result.Error != nil {
			return fmt.Errorf("error upserting package: %v", result.Error)
		}

		// Get latest package version via package
		var latestVersion ent.PackageVersion
		result = tx.Where(ent.PackageVersion{PackageID: pkg.ID}).Order("version DESC").Limit(1).Find(&latestVersion)
		if result.Error != nil {
			return fmt.Errorf("error getting latest package version: %v", result.Error)
		}
		if result.RowsAffected == 0 {
			latestVersion = ent.PackageVersion{
				PackageID: pkg.ID,
				Version:   1,
			}
		}

		// Determine next version
		nextVersion := latestVersion.Version + 1

		// Create a new package version
		pkgVersion := ent.PackageVersion{
			PackageID: pkg.ID,
			Version:   nextVersion,
		}
		result = tx.Create(&pkgVersion)
		if result.Error != nil {
			return fmt.Errorf("error creating package version: %v", result.Error)
		}

		// Save new latest version
		pkg.LatestVersionID = &pkgVersion.ID
		result = tx.Save(&pkg)
		if result.Error != nil {
			return fmt.Errorf("error saving package: %v", result.Error)
		}

	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	// Fetch packages
	var packages []ent.Package
	result := db.Model(&ent.Package{}).Preload("LatestVersion").Limit(10).Find(&packages)
	if result.Error != nil {
		return fmt.Errorf("error fetching packages: %v", result.Error)
	}

	// Print packages
	for _, pck := range packages {
		if pck.LatestVersion == nil {
			fmt.Printf("Package: %s, Version: nil\n", pck.PackageName)
		} else {
			fmt.Printf("Package: %s, Version: %d\n", pck.PackageName, pck.LatestVersion.Version)
		}
	}

	fmt.Println("Uploaded proto files successfully")

	return nil
}

// UploadCommand will upload a set of proto files to the vör service
func UploadCommand() *cli.Command {
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
		},
	}
}
