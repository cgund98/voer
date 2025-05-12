package ctrl

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bufbuild/protocompile/linker"
	v1 "github.com/cgund98/voer/api/v1"
	entity "github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/sqlite"
	"github.com/cgund98/voer/internal/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"gorm.io/gorm"
)

// checkBackwardsCompatible checks if a message is backwards compatible with the latest version of the message.
func checkBackwardsCompatible(ctx context.Context, db *gorm.DB, packageID uint, parsedMsg proto.ParsedMessage) error {

	// Check if message exists
	results := []entity.Message{}
	err := db.Model(&entity.Message{}).Preload("LatestVersion").Where("package_id = ? AND name = ?", packageID, parsedMsg.Name).Limit(1).Find(&results).Error
	if err != nil {
		return fmt.Errorf("failed to check if message exists: %w", err)
	}

	// Check if message version exists
	if len(results) == 0 || results[0].LatestVersion == nil {
		return nil
	}

	msg := results[0]
	msgVersion := msg.LatestVersion

	// Parse schema
	msgSchema, err := proto.DeserializeMessage(msgVersion.SerializedSchema)
	if err != nil {
		return fmt.Errorf("failed to deserialize message schema %s: %w", parsedMsg.Name, err)
	}

	// Check if message version is backwards compatible
	err = proto.ValidateBackwardsCompatibleMessage(ctx, msgSchema, parsedMsg)
	if err != nil {
		return fmt.Errorf("failed to validate backwards compatible message %s: %w", parsedMsg.Name, err)
	}

	return nil
}

// createMessageEntities creates message entities for a given package.
// This includes creating the message and message version entities.
func createMessageEntities(ctx context.Context, tx *gorm.DB, reqPkg *v1.PackageFile, packageID uint, packageVersionID uint, fileContentsMap map[string]string, protoFiles []linker.File) error {

	// Build mapping of msg name to file name
	msgNameToFileNameMap := make(map[string]string)
	for _, file := range protoFiles {
		for _, msg := range proto.ParseMessagesFromFile(file) {
			msgNameToFileNameMap[msg.Name] = filepath.Base(file.Path())
		}
	}

	parsedMsgs := make([]proto.ParsedMessage, 0)
	for _, protoFile := range protoFiles {
		// Ensure package name matches
		if string(protoFile.Package()) != reqPkg.PackageName {
			return fmt.Errorf("package name mismatch: %s != %s", string(protoFile.Package()), reqPkg.PackageName)
		}

		// Parse messages from file
		msgs := proto.ParseMessagesFromFile(protoFile)
		parsedMsgs = append(parsedMsgs, msgs...)
	}

	// Fetch all existing messages for this package
	curMessages := make([]entity.Message, 0)
	err := tx.Model(&entity.Message{}).Where("package_id = ?", packageID).Find(&curMessages).Error
	if err != nil {
		return fmt.Errorf("failed to get current messages: %w", err)
	}

	// Build a lookup of current message names
	curMessageNames := make(map[string]bool)
	for _, msg := range curMessages {
		curMessageNames[msg.Name] = true
	}

	for _, msg := range parsedMsgs {
		// Remove from lookup table
		delete(curMessageNames, msg.Name)

		// Check each message for backwards compatibility
		err := checkBackwardsCompatible(ctx, tx, packageID, msg)
		if err != nil {
			return fmt.Errorf("failed to check backwards compatible message: %w", err)
		}

		// Parse message body
		fileName := msgNameToFileNameMap[msg.Name]
		protoBody, err := proto.ExtractMessageDefinitionByName(fileContentsMap[fileName], msg.Name)
		if err != nil {
			return fmt.Errorf("failed to extract message definition: %w", err)
		}

		// Persist message
		message := entity.Message{
			PackageID: packageID,
			Name:      msg.Name,
			ProtoBody: protoBody,
		}
		result := tx.Where(entity.Message{PackageID: packageID, Name: msg.Name}).FirstOrCreate(&message)
		if result.Error != nil {
			return fmt.Errorf("failed to create message: %w", result.Error)
		}

		// Persist message version
		nextMessageVersion, err := entity.GetNextMessageVersion(tx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to get next message version: %w", err)
		}

		serializedSchema, err := proto.SerializeMessage(msg)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %w", err)
		}

		messageVersion := entity.MessageVersion{
			MessageID:        message.ID,
			Version:          nextMessageVersion,
			ProtoBody:        protoBody,
			SerializedSchema: serializedSchema,
			PackageVersionID: packageVersionID,
		}
		result = tx.Create(&messageVersion)
		if result.Error != nil {
			return fmt.Errorf("failed to create message version: %w", result.Error)
		}

		// Persist latest message version
		message.LatestVersionID = &messageVersion.ID
		result = tx.Save(&message)
		if result.Error != nil {
			return fmt.Errorf("failed to save message: %w", result.Error)
		}
	}

	// Check that no messages were deleted
	for msgName := range curMessageNames {
		return fmt.Errorf("backwards incompatible change: message %s was deleted", msgName)
	}

	return nil
}

// createPackageEntities creates package version entities for a given package.
// This includes creating the package and package version entities.
func createPackageEntities(tx *gorm.DB, reqPkg *v1.PackageFile) (*entity.Package, *entity.PackageVersion, error) {
	// Persist package
	pkg := entity.Package{
		PackageName: reqPkg.PackageName,
	}
	result := tx.Where(entity.Package{PackageName: pkg.PackageName}).FirstOrCreate(&pkg)
	if result.Error != nil {
		return nil, nil, fmt.Errorf("failed to create package: %w", result.Error)
	}

	// Persist package version
	nextPackageVersion, err := entity.GetNextPackageVersion(tx, pkg.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get next package version: %w", err)
	}

	pkgVersion := entity.PackageVersion{
		PackageID: pkg.ID,
		Version:   nextPackageVersion,
	}
	err = tx.Create(&pkgVersion).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create package version: %w", err)
	}

	// Persist latest package version
	pkg.LatestVersionID = &pkgVersion.ID
	result = tx.Save(&pkg)
	if result.Error != nil {
		return nil, nil, fmt.Errorf("failed to save package: %w", result.Error)
	}

	// Persist package version files
	for _, file := range reqPkg.Files {
		pkgVersionFile := entity.PackageVersionFile{
			PackageVersionID: pkgVersion.ID,
			FileName:         file.FileName,
			FileContents:     file.FileContents,
		}

		result = tx.Create(&pkgVersionFile)
		if result.Error != nil {
			return nil, nil, fmt.Errorf("failed to create package version file: %w", result.Error)
		}
	}

	return &pkg, &pkgVersion, nil
}

func CreatePackageVersion(ctx context.Context, db *gorm.DB, req *v1.UploadPackageVersionRequest) (*v1.UploadPackageVersionResponse, error) {
	res := &v1.UploadPackageVersionResponse{}

	_, err := sqlite.WithTx(db, func(tx *gorm.DB) (*entity.Package, error) {

		for _, reqPkg := range req.Packages {
			// Generate list of inputs for proto.ParseStrings
			parseInputs := make([]proto.ParseStringInput, 0)

			for _, file := range reqPkg.Files {
				parseInputs = append(parseInputs, proto.ParseStringInput{
					FileName:     file.FileName,
					FileContents: file.FileContents,
				})
			}

			// Build mapping of file name to file contents
			fileContentsMap := make(map[string]string)
			for _, file := range reqPkg.Files {
				fileContentsMap[file.FileName] = file.FileContents
			}

			// Parse strings into proto files
			protoFiles, err := proto.ParseStrings(ctx, parseInputs...)
			if err != nil {
				return nil, fmt.Errorf("failed to parse proto files: %w", err)
			}

			// Validate no duplicate file names
			err = proto.ValidateNoDuplicateFileNames(ctx, protoFiles)
			if err != nil {
				return nil, fmt.Errorf("failed to validate proto files: %w", err)
			}

			// Create package entities
			pkg, pkgVersion, err := createPackageEntities(tx, reqPkg)
			if err != nil {
				return nil, fmt.Errorf("failed to create package version entities: %w", err)
			}

			res.PackageVersions = append(res.PackageVersions, &v1.PackageVersion{
				Id:        uint64(pkg.ID),
				Version:   uint64(pkgVersion.Version),
				CreatedAt: timestamppb.New(pkg.CreatedAt),
				UpdatedAt: timestamppb.New(pkg.UpdatedAt),
				PackageId: uint64(pkg.ID),
			})

			// Create message entities
			err = createMessageEntities(ctx, tx, reqPkg, pkg.ID, pkgVersion.ID, fileContentsMap, protoFiles)
			if err != nil {
				return nil, err
			}

		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ValidatePackageVersion(ctx context.Context, db *gorm.DB, req *v1.ValidatePackageVersionRequest) (*v1.ValidatePackageVersionResponse, error) {

	for _, reqPkg := range req.Packages {
		// Generate list of inputs for proto.ParseStrings
		parseInputs := make([]proto.ParseStringInput, 0)

		for _, file := range reqPkg.Files {
			parseInputs = append(parseInputs, proto.ParseStringInput{
				FileName:     file.FileName,
				FileContents: file.FileContents,
			})
		}

		// Build mapping of file name to file contents
		fileContentsMap := make(map[string]string)
		for _, file := range reqPkg.Files {
			fileContentsMap[file.FileName] = file.FileContents
		}

		// Parse strings into proto files
		protoFiles, err := proto.ParseStrings(ctx, parseInputs...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proto files: %w", err)
		}

		// Validate no duplicate file names
		err = proto.ValidateNoDuplicateFileNames(ctx, protoFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to validate proto files: %w", err)
		}

		// Parse messages from files
		parsedMsgs := make([]proto.ParsedMessage, 0)
		for _, protoFile := range protoFiles {
			msgs := proto.ParseMessagesFromFile(protoFile)
			parsedMsgs = append(parsedMsgs, msgs...)
		}

		// Check if package exists
		pkgs := make([]entity.Package, 0)
		err = db.Model(&entity.Package{}).Where("package_name = ?", reqPkg.PackageName).Find(&pkgs).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get packages: %w", err)
		}

		if len(pkgs) == 0 {
			return &v1.ValidatePackageVersionResponse{
				IsValid: true,
				Error:   "",
			}, nil
		}
		pkg := pkgs[0]

		// Validate messages
		for _, msg := range parsedMsgs {
			err = checkBackwardsCompatible(ctx, db, pkg.ID, msg)
			if err != nil {
				return &v1.ValidatePackageVersionResponse{
					IsValid: false,
					Error:   err.Error(),
				}, nil
			}
		}

	}

	return &v1.ValidatePackageVersionResponse{
		IsValid: true,
		Error:   "",
	}, nil
}

// GetPackageVersion gets a package version by package name and version.
// Returns the package version and all files in the package version.
func GetPackageVersion(ctx context.Context, db *gorm.DB, req *v1.GetPackageVersionRequest) (*v1.GetPackageVersionResponse, error) {

	// Fetch package
	pkgs := []entity.Package{}
	err := db.Model(&entity.Package{}).Where("package_name = ?", req.PackageName).Find(&pkgs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get packages: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package not found")
	}

	pkg := pkgs[0]

	// Fetch package version
	pkgVersions := []entity.PackageVersion{}
	err = db.Model(&entity.PackageVersion{}).Where("package_id = ? AND version = ?", pkg.ID, req.Version).Find(&pkgVersions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get package versions: %w", err)
	}

	if len(pkgVersions) == 0 {
		return nil, fmt.Errorf("package version not found")
	}

	pkgVer := pkgVersions[0]

	files := []entity.PackageVersionFile{}
	err = db.Model(&entity.PackageVersionFile{}).Where("package_version_id = ?", pkgVer.ID).Find(&files).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get package version files: %w", err)
	}

	res := &v1.GetPackageVersionResponse{
		PackageVersion: &v1.PackageVersion{
			Id:      uint64(pkgVer.ID),
			Version: uint64(pkgVer.Version),
		},
	}

	for _, file := range files {
		res.Files = append(res.Files, &v1.PackageVersionFile{
			Id:               uint64(file.ID),
			ProtoContents:    file.FileContents,
			CreatedAt:        timestamppb.New(file.CreatedAt),
			UpdatedAt:        timestamppb.New(file.UpdatedAt),
			PackageVersionId: uint64(pkgVer.ID),
			FileName:         file.FileName,
		})
	}

	return res, nil

}
