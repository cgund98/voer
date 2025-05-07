package proto

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ParsedMessage struct {
	Name     string
	FullName string
	Fields   []ParsedField
}

type ParsedField struct {
	Name        string
	FullName    string
	Number      int
	Kind        string
	Cardinality string
}

// ParsePath will look for proto files in under a specific path
func ParsePath(ctx context.Context, filePaths ...string) (linker.Files, error) {

	parser := &protocompile.Compiler{
		// You can add ImportPaths if your .proto imports others
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{}),
	}

	// Compile one or more .proto files
	files, err := parser.Compile(ctx, filePaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto file: %w", err)
	}

	return files, nil
}

// ParseString will parse a proto file from a string
func ParseString(ctx context.Context, content string) (linker.File, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "proto")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			fmt.Printf("failed to remove temporary file: %v\n", err)
		}
	}()

	parser := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{}),
	}

	files, err := parser.Compile(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto file: %w", err)
	}
	file := files[0]

	return file, nil

}

// parseMessage will parse a message descriptor into a parsedMessage format
// This is used to compare messages
func parseMessage(message protoreflect.MessageDescriptor) ParsedMessage {

	fields := make([]ParsedField, 0)
	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)

		// Parse kind
		kind := string(field.Kind().String())
		if field.Kind() == protoreflect.MessageKind {
			kind = string(field.Message().FullName())
		}

		// Generate new struct
		fields = append(fields, ParsedField{
			Name:        string(field.Name()),
			FullName:    string(field.FullName()),
			Number:      int(field.Number()),
			Kind:        kind,
			Cardinality: field.Cardinality().String(),
		})
	}

	return ParsedMessage{
		Name:     string(message.Name()),
		FullName: string(message.FullName()),
		Fields:   fields,
	}
}

// ExtractMessageDefinition extracts the message definition from a proto file content
func ExtractMessageDefinitionByName(protoContent string, messageName string) (string, error) {
	// Regular expression to match the message block
	// This regex looks for "message" followed by the name and the content between curly braces
	re := regexp.MustCompile(fmt.Sprintf(`message\s+%s\s*\{[^}]*\}`, regexp.QuoteMeta(messageName)))

	// Find the first match
	match := re.FindString(protoContent)
	if match == "" {
		return "", fmt.Errorf("no message definition found")
	}

	return match, nil
}

// ParseMessagesFromFile will parse all messages from a file
func ParseMessagesFromFile(file linker.File) []ParsedMessage {
	messages := make([]ParsedMessage, 0)
	for i := 0; i < file.Messages().Len(); i++ {
		message := file.Messages().Get(i)
		messages = append(messages, parseMessage(message))
	}
	return messages
}

// GetFieldByNumber will return the field with the given number
func GetFieldByNumber(fields []ParsedField, number int) *ParsedField {
	for _, field := range fields {
		if field.Number == number {
			return &field
		}
	}
	return nil
}

// GetMessageByName will return the message with the given full name
func GetMessageByName(messages []ParsedMessage, fullName string) *ParsedMessage {
	for _, message := range messages {
		if message.FullName == fullName {
			return &message
		}
	}
	return nil
}
