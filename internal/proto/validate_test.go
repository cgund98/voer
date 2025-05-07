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

// TestValidateBackwardsCompatibileMessageEqual tests that a message is backwards compatible with itself
func TestValidateBackwardsCompatibileMessagesEqual(t *testing.T) {
	fileContent := `
	syntax = "proto3";

	package helloworld;

	message Greeting {
		string message = 1;
	}
	`

	ctx := context.Background()
	file := createTempProto(t, ctx, fileContent)

	err := ValidateBackwardsCompatibileMessages(ctx, file.Messages(), file.Messages())
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

// TestValidateBackwardsCompatibileMessageEqual tests that a message is backwards compatible with itself
func TestValidateBackwardsCompatibileMessagesAddedField(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

func TestValidateBackwardsCompatibileMessagesAddedNestedField(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err != nil {
		t.Fatalf("Failed to validate backwards compatible message: %v", err)
	}
}

func TestValidateBackwardsCompatibileMessagesRemovedField(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesRemovedNestedField(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Address: field 'city' was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesChangedName(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed name to 'name' which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesChangedType(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed type from string to int32 which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesChangedNestedType(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'address' changed type from helloworld.Address to helloworld.NewAddress which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesChangedCardinality(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting: field 'age' changed cardinality from optional to repeated which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateBackwardsCompatibileMessagesRemovedMessage(t *testing.T) {
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

	err := ValidateBackwardsCompatibileMessages(ctx, prevFile.Messages(), latestFile.Messages())
	if err == nil {
		t.Fatalf("Expected error for removed field")
	}

	expectedError := "message helloworld.Greeting was removed which breaks backwards compatibility"
	if err.Error() != expectedError {
		t.Fatalf("Expected error: %s, got: %s", expectedError, err.Error())
	}
}

func TestValidateUniquePackageUniqueNames(t *testing.T) {
	firstContent := `
	syntax = "proto3";

	package helloworld;
	`

	secondContent := `
	syntax = "proto3";

	package goodbyeworld;
	`

	ctx := context.Background()
	firstFile := createTempProto(t, ctx, firstContent)
	secondFile := createTempProto(t, ctx, secondContent)

	err := ValidateUniquePackage(ctx, linker.Files{firstFile, secondFile})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestValidateUniquePackageDuplicateNames(t *testing.T) {
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

	err := ValidateUniquePackage(ctx, linker.Files{firstFile, secondFile})
	if err == nil {
		t.Fatalf("Expected error for duplicate package name")
	}
}
