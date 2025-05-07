package proto

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/bufbuild/protocompile/linker"
)

// ValidateBackwardsCompatibleMessage checks if a message descriptor is backwards compatible with another
func ValidateBackwardsCompatibleMessage(ctx context.Context, previous, latest ParsedMessage) error {

	// Compare fields between previous and latest versions
	prevFields := previous.Fields

	// Check each field in the previous version
	for i := 0; i < len(prevFields); i++ {
		prevField := prevFields[i]
		latestField := GetFieldByNumber(latest.Fields, prevField.Number)

		// Field was removed in latest version
		if latestField == nil {
			return fmt.Errorf("field '%s' was removed which breaks backwards compatibility", prevField.Name)
		}

		// Check name changes
		if prevField.FullName != latestField.FullName {
			return fmt.Errorf("field '%s' changed name to '%s' which breaks backwards compatibility",
				prevField.Name, latestField.Name)
		}

		// Check field type changes
		if prevField.Kind != latestField.Kind {
			return fmt.Errorf("field '%s' changed type from %v to %v which breaks backwards compatibility",
				prevField.Name, prevField.Kind, latestField.Kind)
		}

		// Check cardinality changes (required/optional/repeated)
		if prevField.Cardinality != latestField.Cardinality {
			return fmt.Errorf("field '%s' changed cardinality from %v to %v which breaks backwards compatibility",
				prevField.Name, prevField.Cardinality, latestField.Cardinality)
		}
	}

	return nil
}

// ValidateBackwardsCompatibleMessages checks if a set of messages are backwards compatible with another
func ValidateBackwardsCompatibleMessages(ctx context.Context, prevMessages, latestMessages []ParsedMessage) error {

	for i := 0; i < len(prevMessages); i++ {
		prevMessage := prevMessages[i]
		latestMessage := GetMessageByName(latestMessages, prevMessage.FullName)

		if latestMessage == nil {
			return fmt.Errorf("message %s was removed which breaks backwards compatibility", prevMessage.FullName)
		}

		if err := ValidateBackwardsCompatibleMessage(ctx, prevMessage, *latestMessage); err != nil {
			return fmt.Errorf("message %s: %w", prevMessage.FullName, err)
		}
	}

	return nil
}

func getParentPath(file linker.File) string {
	filePath := string(file.Path())
	parentPath := filepath.Dir(filePath)
	return parentPath
}

// ValidatePackagesInSameDirectory will make sure files sharing the same package name are in the same directory
func ValidatePackagesInSameDirectory(ctx context.Context, protoFiles linker.Files) error {

	packageNames := make(map[string]string)
	for _, protoFile := range protoFiles {
		if curPath, ok := packageNames[string(protoFile.Package())]; ok {
			fmt.Println(getParentPath(protoFile))
			if curPath != getParentPath(protoFile) {
				return fmt.Errorf("proto file %s has the same package name as %s but is in a different directory", protoFile.Path(), curPath)
			}
		}
		packageNames[string(protoFile.Package())] = getParentPath(protoFile)
	}

	return nil
}
