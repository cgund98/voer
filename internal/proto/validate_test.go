package proto

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bufbuild/protocompile/linker"
)

// createTempProto creates a temporary proto file and returns the file descriptor
func createTempProto(t *testing.T, ctx context.Context, content string) linker.File {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.proto")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Parse the proto file
	files, err := ParsePath(ctx, filePath)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	return files[0]
}

// TestValidateBackwardsCompatibleMessageEqual tests that a message is backwards compatible with itself
func TestValidateBackwardsCompatibleMessagesEqual(t *testing.T) {
	fileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
	}
	`

	ctx := context.Background()
	file := createTempProto(t, ctx, fileContent)

	messages := ParseMessagesFromFile(file)

	err := ValidateBackwardsCompatibleMessages(ctx, messages, messages)
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

// TestValidateBackwardsCompatibleMessageEqual tests that a message is backwards compatible with itself
func TestValidateBackwardsCompatibleMessagesAddedField(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}

	message Response {
		string message = 1;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

func TestValidateBackwardsCompatibleMessagesAddedNestedField(t *testing.T) {
	prevFileContent := `
		syntax = "proto3";

		package helloworld;

		message Address {
			string street = 1;
		}

		message Greeting {
			string message = 1;
			Address address = 2;
		}
		`

	latestFileContent := `
		syntax = "proto3";

		package helloworld;

		message Address {
			string street = 1;
			string city = 2;
		}

		message Greeting {
			string message = 1;
			Address address = 2;
		}
		`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

func TestValidateBackwardsCompatibleMessagesRemovedField(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesRemovedNestedField(t *testing.T) {
	prevFileContent := `
		syntax = "proto3";

		package helloworld;

		message Greeting {
			string message = 1;
			Address address = 2;
		}

		message Address {
			string street = 1;
			string city = 2;
		}
		`

	latestFileContent := `
		syntax = "proto3";

		package helloworld;

		message Greeting {
			string message = 1;
			Address address = 2;
		}

		message Address {
			string street = 1;
		}
		`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Address: field 'city' was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesChangedName(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string name = 2;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed name to 'name' which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesChangedType(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		int32 age = 2;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed type from string to int32 which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesChangedNestedType(t *testing.T) {
	prevFileContent := `
		syntax = "proto3";

		package helloworld;

		message Greeting {
			string message = 1;
			Address address = 2;
		}

		message Address {
			string street = 1;
		}
		`

	latestFileContent := `
		syntax = "proto3";

		package helloworld;

		message Greeting {
			string message = 1;
			NewAddress address = 2;
		}

		message Address {
			string street = 1;
		}

		message NewAddress {
			string street = 1;
			string city = 2;
		}
		`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'address' changed type from helloworld.Address to helloworld.NewAddress which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesChangedCardinality(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		repeated string age = 2;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed cardinality from optional to repeated which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibleMessagesRemovedMessage(t *testing.T) {
	prevFileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
		string age = 2;
	}
	`

	latestFileContent := `
	syntax = "proto3";

	package helloworld;

	message Response {
		string message = 1;
	}
	`

	ctx := context.Background()
	prevFile := createTempProto(t, ctx, prevFileContent)
	latestFile := createTempProto(t, ctx, latestFileContent)

	prevMessages := ParseMessagesFromFile(prevFile)
	latestMessages := ParseMessagesFromFile(latestFile)

	err := ValidateBackwardsCompatibleMessages(ctx, prevMessages, latestMessages)
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidatePackagesInSameDirectory(t *testing.T) {
	firstContent := `
	syntax = "proto3";

	package helloworld;
	`

	secondContent := `
	syntax = "proto3";

	package helloworld;
	`

	ctx := context.Background()

	tempDir := t.TempDir()
	firstFilePath := filepath.Join(tempDir, "request.proto")
	secondFilePath := filepath.Join(tempDir, "response.proto")

	err := os.WriteFile(firstFilePath, []byte(firstContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = os.WriteFile(secondFilePath, []byte(secondContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	firstFiles, err := ParsePath(ctx, firstFilePath)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}
	secondFiles, err := ParsePath(ctx, secondFilePath)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	err = ValidatePackagesInSameDirectory(ctx, linker.Files{firstFiles[0], secondFiles[0]})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestValidatePackagesInSameDirectoryDifferentDirectories(t *testing.T) {
	firstContent := `
	syntax = "proto3";

	package helloworld;
	`

	secondContent := `
	syntax = "proto3";

	package helloworld;
	`

	ctx := context.Background()
	firstFile := createTempProto(t, ctx, firstContent)
	secondFile := createTempProto(t, ctx, secondContent)

	err := ValidatePackagesInSameDirectory(ctx, linker.Files{firstFile, secondFile})
	if err == nil {
		t.Fatalf("Expected error for different directories")
	}
}

func TestValidateNoDuplicateFileNames(t *testing.T) {
	content := `
	syntax = "proto3";

	package helloworld;
	`

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	filePath1 := filepath.Join(dir1, "request.proto")
	filePath2 := filepath.Join(dir2, "request.proto")

	err := os.WriteFile(filePath1, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = os.WriteFile(filePath2, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	ctx := context.Background()
	files, err := ParsePath(ctx, filePath1, filePath2)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	err = ValidateNoDuplicateFileNames(ctx, files)
	if err == nil {
		t.Fatalf("Expected error for duplicate file names")
	}
}

func TestValidateNoDuplicateFileNamesDifferentNames(t *testing.T) {
	content := `
	syntax = "proto3";

	package helloworld;
	`

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	filePath1 := filepath.Join(dir1, "request.proto")
	filePath2 := filepath.Join(dir2, "response.proto")

	err := os.WriteFile(filePath1, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	err = os.WriteFile(filePath2, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	ctx := context.Background()
	files, err := ParsePath(ctx, filePath1, filePath2)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	err = ValidateNoDuplicateFileNames(ctx, files)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
