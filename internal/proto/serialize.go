package proto

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// SerializeMessage will serialize a message into a string.
// We'll use a json format for the serialization.
func SerializeMessage(message ParsedMessage) string {

	json, err := json.Marshal(message)
	if err != nil {
		return ""
	}

	return string(json)
}

// DeserializeMessage will deserialize a message from a string.
func DeserializeMessage(serializedMessage string) ParsedMessage {
	var message ParsedMessage
	err := json.Unmarshal([]byte(serializedMessage), &message)
	if err != nil {
		return ParsedMessage{}
	}

	return message
}

// DumpProtoMessage will attempt to reconstruct a .proto definition from a message.
// This is not always possible, but it's a good way to debug the message.
// It will mainly be used for viewing in the UI.
func DumpProtoMessage(message protoreflect.MessageDescriptor) string {
	// Start with message definition
	output := ""

	// Add package name
	output += fmt.Sprintf("package %s;\n", message.FullName().Parent().Name())

	// Add message definition
	output += fmt.Sprintf("message %s {\n", message.Name())

	// Add each field
	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)
		output += fmt.Sprintf("  %s %s = %d;\n", field.Cardinality(), field.Kind(), field.Number())
	}

	// Close message definition
	output += "}"

	return output
}
