package util

import "testing"

func TestValidateMf2Success(t *testing.T) {
	doc := Mf2Document{
		Type:       []string{"h-entry"},
		Properties: map[string][]any{"name": []any{"hello"}},
	}

	if err := ValidateMf2(doc); err != nil {
		t.Fatalf("expected document to be valid: %v", err)
	}
}

func TestValidateMf2Errors(t *testing.T) {
	cases := []Mf2Document{
		{Type: nil, Properties: map[string][]any{"name": []any{"x"}}},
		{Type: []string{""}, Properties: map[string][]any{"name": []any{"x"}}},
		{Type: []string{"h-entry"}, Properties: nil},
		{Type: []string{"h-entry"}, Properties: map[string][]any{"": []any{"x"}}},
		{Type: []string{"h-entry"}, Properties: map[string][]any{"name": []any{123}}},
	}

	for i, doc := range cases {
		if err := ValidateMf2(doc); err == nil {
			t.Fatalf("expected case %d to be invalid", i)
		}
	}
}

func TestValidateMf2Embedded(t *testing.T) {
	nested := Mf2Document{Type: []string{"h-card"}, Properties: map[string][]any{"name": []any{"Alice"}}}
	doc := Mf2Document{Type: []string{"h-entry"}, Properties: map[string][]any{"author": []any{nested}}}

	if err := ValidateMf2(doc); err != nil {
		t.Fatalf("expected embedded mf2 to be valid: %v", err)
	}

	badNested := Mf2Document{Type: []string{}, Properties: map[string][]any{"name": []any{"Bob"}}}
	doc.Properties["author"] = []any{badNested}
	if err := ValidateMf2(doc); err == nil {
		t.Fatalf("expected invalid embedded mf2 to be rejected")
	}
}
