package post

import "testing"

func TestGetDeletionsArray(t *testing.T) {
	input := map[string]any{"delete": []any{"category", "photo"}}
	result, err := getDeletions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	props, ok := result.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", result)
	}
	if len(props) != 2 || props[0] != "category" || props[1] != "photo" {
		t.Fatalf("unexpected props: %#v", props)
	}
}

func TestGetDeletionsMap(t *testing.T) {
	input := map[string]any{"delete": map[string]any{"category": []any{"foo", "bar"}, "tags": "one"}}
	result, err := getDeletions(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	props, ok := result.(map[string][]any)
	if !ok {
		t.Fatalf("expected map[string][]any, got %T", result)
	}
	if len(props["category"]) != 2 || len(props["tags"]) != 1 {
		t.Fatalf("unexpected values: %#v", props)
	}
}

func TestGetDeletionsInvalid(t *testing.T) {
	input := map[string]any{"delete": 123}
	if _, err := getDeletions(input); err == nil {
		t.Fatalf("expected error for invalid delete type")
	}
}
