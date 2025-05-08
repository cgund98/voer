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
		return fmt.Errorf("failed to deserialize message schema: %w", err)
	}

	// Check if message version is backwards compatible
	err = proto.ValidateBackwardsCompatibleMessage(ctx, msgSchema, parsedMsg)
	if err != nil {
		return fmt.Errorf("failed to validate backwards compatible message: %w", err)
	}

	return nil
}

// createMessageEntities creates message entities for a given package.
// This includes creating the message and message version entities.
func createMessageEntities(ctx context.Context, tx *gorm.DB, reqPkg *v1.PackageFile, packageID uint, fileContentsMap map[string]string, protoFiles []linker.File) error {

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

	for _, msg := range parsedMsgs {
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

		// Log new message version
		fmt.Printf("%s v%d\n", msg.Name, nextMessageVersion)

	}

	return nil
}

// createPackageEntities creates package version entities for a given package.
// This includes creating the package and package version entities.
func createPackageEntities(tx *gorm.DB, reqPkg *v1.PackageFile) (*entity.Package, error) {
	// Persist package
	pkg := entity.Package{
		PackageName: reqPkg.PackageName,
	}
	result := tx.Where(entity.Package{PackageName: pkg.PackageName}).FirstOrCreate(&pkg)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create package: %w", result.Error)
	}

	// Persist package version
	nextPackageVersion, err := entity.GetNextPackageVersion(tx, pkg.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next package version: %w", err)
	}

	pkgVersion := entity.PackageVersion{
		PackageID: pkg.ID,
		Version:   nextPackageVersion,
	}
	err = tx.Create(&pkgVersion).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create package version: %w", err)
	}

	// Persist latest package version
	pkg.LatestVersionID = &pkgVersion.ID
	result = tx.Save(&pkg)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to save package: %w", result.Error)
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
			return nil, fmt.Errorf("failed to create package version file: %w", result.Error)
		}
	}

	return &pkg, nil
}

func CreatePackageVersion(ctx context.Context, db *gorm.DB, req *v1.UploadPackageVersionRequest) (*v1.UploadPackageVersionResponse, error) {

	// Generate list of inputs for proto.ParseStrings
	parseInputs := make([]proto.ParseStringInput, 0)

	_, err := sqlite.WithTx(db, func(tx *gorm.DB) (*entity.Package, error) {

		for _, reqPkg := range req.Packages {

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
			pkg, err := createPackageEntities(tx, reqPkg)
			if err != nil {
				return nil, fmt.Errorf("failed to create package version entities: %w", err)
			}

			// Create message entities
			err = createMessageEntities(ctx, tx, reqPkg, pkg.ID, fileContentsMap, protoFiles)
			if err != nil {
				return nil, err
			}

		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return &v1.UploadPackageVersionResponse{}, nil
}
