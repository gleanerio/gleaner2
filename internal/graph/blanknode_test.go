package graph

import (
	"testing"
)

func TestExtractSchemaType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<https://schema.org/Person>", "Person"},
		{"<http://schema.org/Dataset>", "Dataset"},
		{"<https://schema.org/PropertyValue>", "PropertyValue"},
		{"<http://www.w3.org/ns/prov#Activity>", "Activity"},
		{"<https://example.org/types/Custom>", "Custom"},
	}

	for _, tt := range tests {
		got := extractSchemaType(tt.input)
		if got != tt.expected {
			t.Errorf("extractSchemaType(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestExtractLiteralValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello world"`, "hello world"},
		{`"value"^^<http://www.w3.org/2001/XMLSchema#string>`, "value"},
		{`<https://example.org/file.csv>`, "https://example.org/file.csv"},
		{`"escaped \"quotes\""`, `escaped "quotes"`},
	}

	for _, tt := range tests {
		got := extractLiteralValue(tt.input)
		if got != tt.expected {
			t.Errorf("extractLiteralValue(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGenerateDeterministicID_ContentUrl(t *testing.T) {
	info := &blankNodeInfo{
		rdfType:    "Dataset",
		contentUrl: "https://example.org/data.csv",
	}
	id := generateDeterministicID(info)
	if id == "" {
		t.Fatal("expected non-empty ID for node with contentUrl")
	}
	if got := generateDeterministicID(info); got != id {
		t.Errorf("deterministic ID changed: first=%q, second=%q", id, got)
	}
	if !contains(id, "/Dataset/") {
		t.Errorf("expected schema type in URI, got %q", id)
	}
}

func TestGenerateDeterministicID_NameValue(t *testing.T) {
	info := &blankNodeInfo{
		rdfType: "PropertyValue",
		name:    "color",
		value:   "blue",
	}
	id := generateDeterministicID(info)
	if id == "" {
		t.Fatal("expected non-empty ID for node with name+value")
	}
	if !contains(id, "/PropertyValue/") {
		t.Errorf("expected schema type in URI, got %q", id)
	}
}

func TestGenerateDeterministicID_Title(t *testing.T) {
	info := &blankNodeInfo{
		rdfType: "CreativeWork",
		title:   "My Document",
	}
	id := generateDeterministicID(info)
	if id == "" {
		t.Fatal("expected non-empty ID for node with title")
	}
	if !contains(id, "/CreativeWork/") {
		t.Errorf("expected schema type in URI, got %q", id)
	}
}

func TestGenerateDeterministicID_NameOnly(t *testing.T) {
	info := &blankNodeInfo{
		rdfType: "Organization",
		name:    "ACME Corp",
	}
	id := generateDeterministicID(info)
	if id == "" {
		t.Fatal("expected non-empty ID for node with name")
	}
	if !contains(id, "/Organization/") {
		t.Errorf("expected schema type in URI, got %q", id)
	}
}

func TestGenerateDeterministicID_NoProperties(t *testing.T) {
	info := &blankNodeInfo{rdfType: "Thing"}
	id := generateDeterministicID(info)
	if id != "" {
		t.Errorf("expected empty ID for node with no qualifying properties, got %q", id)
	}
}

func TestGenerateDeterministicID_NilInfo(t *testing.T) {
	id := generateDeterministicID(nil)
	if id != "" {
		t.Errorf("expected empty ID for nil info, got %q", id)
	}
}

func TestGenerateDeterministicID_DefaultType(t *testing.T) {
	info := &blankNodeInfo{
		name: "test",
	}
	id := generateDeterministicID(info)
	if !contains(id, "/Thing/") {
		t.Errorf("expected default type 'Thing' in URI, got %q", id)
	}
}

func TestGenerateDeterministicID_Priority(t *testing.T) {
	// contentUrl should take priority over name+value, title, name
	infoAll := &blankNodeInfo{
		rdfType:    "Dataset",
		contentUrl: "https://example.org/data.csv",
		name:       "test",
		value:      "val",
		title:      "title",
	}
	infoContentOnly := &blankNodeInfo{
		rdfType:    "Dataset",
		contentUrl: "https://example.org/data.csv",
	}
	if generateDeterministicID(infoAll) != generateDeterministicID(infoContentOnly) {
		t.Error("contentUrl should take priority — IDs should match regardless of other properties")
	}

	// name+value should take priority over title and name-alone
	infoNV := &blankNodeInfo{
		rdfType: "PropertyValue",
		name:    "color",
		value:   "blue",
		title:   "ignored",
	}
	infoNVOnly := &blankNodeInfo{
		rdfType: "PropertyValue",
		name:    "color",
		value:   "blue",
	}
	if generateDeterministicID(infoNV) != generateDeterministicID(infoNVOnly) {
		t.Error("name+value should take priority over title")
	}
}

func TestCollectBlankNodeProperties(t *testing.T) {
	lines := []string{
		`_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Dataset> .`,
		`_:b0 <https://schema.org/name> "Test Dataset" .`,
		`_:b0 <https://schema.org/contentUrl> <https://example.org/data.csv> .`,
		`_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/PropertyValue> .`,
		`_:b1 <https://schema.org/name> "color" .`,
		`_:b1 <https://schema.org/value> "blue" .`,
	}

	props := collectBlankNodeProperties(lines)

	b0 := props["_:b0"]
	if b0 == nil {
		t.Fatal("expected properties for _:b0")
	}
	if b0.rdfType != "Dataset" {
		t.Errorf("_:b0 type = %q, want Dataset", b0.rdfType)
	}
	if b0.contentUrl != "https://example.org/data.csv" {
		t.Errorf("_:b0 contentUrl = %q, want https://example.org/data.csv", b0.contentUrl)
	}
	if b0.name != "Test Dataset" {
		t.Errorf("_:b0 name = %q, want 'Test Dataset'", b0.name)
	}

	b1 := props["_:b1"]
	if b1 == nil {
		t.Fatal("expected properties for _:b1")
	}
	if b1.rdfType != "PropertyValue" {
		t.Errorf("_:b1 type = %q, want PropertyValue", b1.rdfType)
	}
	if b1.name != "color" {
		t.Errorf("_:b1 name = %q, want color", b1.name)
	}
	if b1.value != "blue" {
		t.Errorf("_:b1 value = %q, want blue", b1.value)
	}
}

func TestSkolemization_Deterministic(t *testing.T) {
	nq := `_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Dataset> .
_:b0 <https://schema.org/name> "Test Dataset" .
_:b0 <https://schema.org/contentUrl> <https://example.org/data.csv> .
<https://example.org/record> <https://schema.org/about> _:b0 .
`
	result1, err := Skolemization(nq, "test")
	if err != nil {
		t.Fatal(err)
	}
	result2, err := Skolemization(nq, "test")
	if err != nil {
		t.Fatal(err)
	}
	if result1 != result2 {
		t.Errorf("Skolemization should be deterministic for nodes with properties.\nFirst:  %s\nSecond: %s", result1, result2)
	}
	if contains(result1, "_:b0") {
		t.Error("blank node _:b0 should have been replaced")
	}
	if !contains(result1, "gleaner.io/id/Dataset/") {
		t.Errorf("expected content-based URI with schema type, got:\n%s", result1)
	}
}

func TestSkolemization_FallbackToXID(t *testing.T) {
	// Blank node with no qualifying properties — should fall back to XID
	nq := `_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Thing> .
<https://example.org/record> <https://schema.org/about> _:b0 .
`
	result, err := Skolemization(nq, "test")
	if err != nil {
		t.Fatal(err)
	}
	if contains(result, "_:b0") {
		t.Error("blank node _:b0 should have been replaced")
	}
	if !contains(result, "gleaner.io/xid/genid/") {
		t.Errorf("expected XID fallback URI, got:\n%s", result)
	}
}

func TestSkolemization_MultipleBlankNodes(t *testing.T) {
	nq := `_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Dataset> .
_:b0 <https://schema.org/name> "Dataset A" .
_:b1 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://schema.org/Person> .
_:b1 <https://schema.org/name> "Alice" .
_:b0 <https://schema.org/author> _:b1 .
`
	result, err := Skolemization(nq, "test")
	if err != nil {
		t.Fatal(err)
	}
	if contains(result, "_:b0") || contains(result, "_:b1") {
		t.Error("all blank nodes should have been replaced")
	}
	if !contains(result, "/Dataset/") {
		t.Error("expected Dataset type in URI")
	}
	if !contains(result, "/Person/") {
		t.Error("expected Person type in URI")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
