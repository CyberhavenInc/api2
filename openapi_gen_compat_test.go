package api2_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/starius/api2"
	"github.com/starius/api2/example"
)

func TestGenerateOpenApiSpecMatchesGolden(t *testing.T) {
	got := generateExampleOpenAPI(t)
	want, err := os.ReadFile(filepath.Join("example", "openapi", "openapi.json"))
	if err != nil {
		t.Fatalf("read golden OpenAPI spec: %v", err)
	}
	if !reflect.DeepEqual(parseJSON(t, got), parseJSON(t, want)) {
		t.Fatalf("generated OpenAPI spec no longer matches the committed fixture semantically")
	}
}

func TestGenerateOpenApiSpecValidatesLegacyShape(t *testing.T) {
	generated := generateExampleOpenAPI(t)

	loader := spec.NewLoader()
	doc, err := loader.LoadFromData(generated)
	if err != nil {
		t.Fatalf("load generated OpenAPI spec: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validate generated OpenAPI spec: %v", err)
	}
}

func TestGenerateOpenApiSpecLegacyCompatibility(t *testing.T) {
	generated := generateExampleOpenAPI(t)
	want, err := os.ReadFile(filepath.Join("example", "openapi", "openapi.json"))
	if err != nil {
		t.Fatalf("read golden OpenAPI spec: %v", err)
	}

	gotLegacy := decodeLegacyDoc(t, generated)
	wantLegacy := decodeLegacyDoc(t, want)
	if !reflect.DeepEqual(gotLegacy, wantLegacy) {
		t.Fatalf("generated OpenAPI spec is not backwards compatible with the legacy JSON shape")
	}
}

func generateExampleOpenAPI(t *testing.T) []byte {
	t.Helper()

	outDir := t.TempDir()
	api2.GenerateOpenApiSpec(&api2.TypesGenConfig{
		OutDir: outDir,
		Routes: []interface{}{example.GetRoutes},
		Types: []interface{}{
			&example.EchoRequest{},
			&example.CustomType2{},
		},
	})

	content, err := os.ReadFile(filepath.Join(outDir, "openapi.json"))
	if err != nil {
		t.Fatalf("read generated OpenAPI spec: %v", err)
	}
	return content
}

func parseJSON(t *testing.T, data []byte) any {
	t.Helper()

	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	return value
}

func decodeLegacyDoc(t *testing.T, data []byte) legacyDoc {
	t.Helper()

	var doc legacyDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode legacy OpenAPI view: %v", err)
	}
	if len(doc.Paths) == 0 {
		t.Fatal("legacy compatibility broken: paths is empty")
	}
	if len(doc.Components.Schemas) == 0 {
		t.Fatal("legacy compatibility broken: components.schemas is empty")
	}
	if len(doc.Components.RequestBodies) == 0 {
		t.Fatal("legacy compatibility broken: components.requestBodies is empty")
	}
	return doc
}

type legacyDoc struct {
	OpenAPI    string           `json:"openapi"`
	Paths      map[string]any   `json:"paths"`
	Components legacyComponents `json:"components"`
}

type legacyComponents struct {
	Schemas       map[string]legacySchemaRef      `json:"schemas"`
	RequestBodies map[string]legacyRequestBodyRef `json:"requestBodies"`
}

type legacyRequestBodyRef struct {
	Ref     string                     `json:"$ref,omitempty"`
	Content map[string]legacyMediaType `json:"content,omitempty"`
}

type legacyMediaType struct {
	Schema *legacySchemaRef `json:"schema,omitempty"`
}

type legacySchemaRef struct {
	Ref                  string                     `json:"$ref,omitempty"`
	Type                 string                     `json:"type,omitempty"`
	Format               string                     `json:"format,omitempty"`
	Description          string                     `json:"description,omitempty"`
	Nullable             bool                       `json:"nullable,omitempty"`
	Enum                 []any                      `json:"enum,omitempty"`
	Properties           map[string]legacySchemaRef `json:"properties,omitempty"`
	Items                *legacySchemaRef           `json:"items,omitempty"`
	AllOf                []legacySchemaRef          `json:"allOf,omitempty"`
	OneOf                []legacySchemaRef          `json:"oneOf,omitempty"`
	AnyOf                []legacySchemaRef          `json:"anyOf,omitempty"`
	AdditionalProperties *legacySchemaRef           `json:"additionalProperties,omitempty"`
}
