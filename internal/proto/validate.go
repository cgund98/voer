package proto

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValidateBackwardsCompatibileMessage checks if a message descriptor is backwards compatible with another
func ValidateBackwardsCompatibileMessage(ctx context.Context, previous, latest protoreflect.MessageDescriptor) error {

	// Compare fields between previous and latest versions
	prevFields := previous.Fields()
	latestFields := latest.Fields()

	// Check each field in the previous version
	for i := 0; i < prevFields.Len(); i++ {
		prevField := prevFields.Get(i)
		latestField := latestFields.ByNumber(prevField.Number())

		// Field was removed in latest version
		if latestField == nil {
			return fmt.Errorf("field '%s' was removed which breaks backwards compatibility", prevField.Name())
		}

		// Check name changes
		if prevField.FullName() != latestField.FullName() {
			return fmt.Errorf("field '%s' changed name to '%s' which breaks backwards compatibility",
				prevField.Name(), latestField.Name())
		}

		// Check field type changes
		if prevField.Kind() != latestField.Kind() {
			return fmt.Errorf("field '%s' changed type from %v to %v which breaks backwards compatibility",
				prevField.Name(), prevField.Kind(), latestField.Kind())
		} else if prevField.Kind() == protoreflect.MessageKind {
			if prevField.Message().FullName() != latestField.Message().FullName() {
				return fmt.Errorf("field '%s' changed type from %s to %s which breaks backwards compatibility",
					prevField.Name(), prevField.Message().FullName(), latestField.Message().FullName())
			}
		}

		// Check cardinality changes (required/optional/repeated)
		if prevField.Cardinality() != latestField.Cardinality() {
			return fmt.Errorf("field '%s' changed cardinality from %v to %v which breaks backwards compatibility",
				prevField.Name(), prevField.Cardinality(), latestField.Cardinality())
		}
	}

	return nil
}

// ValidateBackwardsCompatibileMessages checks if a set of messages are backwards compatible with another
func ValidateBackwardsCompatibileMessages(ctx context.Context, prevMessages, latestMessages protoreflect.MessageDescriptors) error {

	for i := 0; i < prevMessages.Len(); i++ {
		prevMessage := prevMessages.Get(i)
		latestMessage := latestMessages.ByName(prevMessage.Name())

		if latestMessage == nil {
			return fmt.Errorf("message %s was removed which breaks backwards compatibility", prevMessage.FullName())
		}

		if err := ValidateBackwardsCompatibileMessage(ctx, prevMessage, latestMessage); err != nil {
			return fmt.Errorf("message %s: %w", prevMessage.FullName(), err)
		}
	}

	return nil
}
