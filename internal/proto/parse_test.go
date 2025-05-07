package proto

import "testing"

func TestExtractMessageDefinitionByName(t *testing.T) {
	content := `
	syntax = "proto3";

	package helloworld;

	message GreetingRequest {
		string message = 1;
	}

	message GreetingResponse {
		string message = 1;
	}
	`

	message, err := ExtractMessageDefinitionByName(content, "GreetingRequest")
	if err != nil {
		t.Fatalf("error extracting message definition: %v", err)
	}

	expected := `message GreetingRequest {
		string message = 1;
	}`

	if message != expected {
		t.Fatalf("expected message definition: %v, got: %v", expected, message)
	}
}

func TestExtractMessageDefinitionByNameNotFound(t *testing.T) {
	content := `
	syntax = "proto3";

	package helloworld;
	`

	_, err := ExtractMessageDefinitionByName(content, "GreetingRequest")
	if err == nil {
		t.Fatalf("expected error for not found message")
	}
}
