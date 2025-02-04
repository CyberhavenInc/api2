package api2

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/starius/api2/typegen"
)

func GenerateOpenApiSpec(options *TypesGenConfig) {
	_ = os.RemoveAll(filepath.Join(options.OutDir, "openapi.json"))
	err := os.MkdirAll(options.OutDir, os.ModePerm)
	panicIf(err)
	typesFile, err := os.OpenFile(filepath.Join(options.OutDir, "openapi.json"), os.O_WRONLY|os.O_CREATE, 0755)
	panicIf(err)
	parser := typegen.NewParser()
	parser.CustomParse = CustomParse
	allRoutes := []Route{}
	for _, getRoutes := range options.Routes {
		genValue := reflect.ValueOf(getRoutes)
		serviceArg := reflect.New(genValue.Type().In(0)).Elem()
		routesValues := genValue.Call([]reflect.Value{serviceArg})
		routes := routesValues[0].Interface().([]Route)
		allRoutes = append(allRoutes, routes...)
	}

	parser.ParseRaw(options.Types...)
	swag := spec.T{
		OpenAPI: "3.0.0",
		Info: &spec.Info{
			Version: "3.0.0",
		},
		Paths: spec.Paths{},
		Components: &spec.Components{
			RequestBodies: spec.RequestBodies{},
		},
	}

	genOpenApiRoutes(typesFile, allRoutes, parser, options, &swag)
	typegen.PrintSwagger(parser, &swag)
	content, err := json.MarshalIndent(swag, "", " ")
	panicIf(err)
	_, err = typesFile.Write(content)
	panicIf(err)
}

func genOpenApiRoutes(w io.Writer, routes []Route, p *typegen.Parser, options *TypesGenConfig, swagger *spec.T) {
	type routeDef struct {
		Method      string
		Path        string
		ReqType     string
		ResType     string
		Handler     interface{}
		FnInfo      FnInfo
		TypeInfoReq string
		TypeInfoRes string
	}
	m := map[string]map[string][]routeDef{}
OUTER:
	for _, route := range routes {
		handler := route.Handler
		if f, ok := handler.(funcer); ok {
			handler = f.Func()
		}

		handlerVal := reflect.ValueOf(handler)
		handlerType := handlerVal.Type()
		req := reflect.TypeOf(reflect.New(handlerType.In(1)).Elem().Interface()).Elem()
		response := reflect.TypeOf(reflect.New(handlerType.Out(0)).Elem().Interface()).Elem()
		fnInfo := GetFnInfo(route.Handler)
		for _, v := range options.Blacklist {
			if Matches(&v, fnInfo.PkgName, fnInfo.StructName, fnInfo.Method) {
				continue OUTER
			}
		}
		p.Parse(req, response)
		TypeInfoReq, err := serializeTypeInfo(prepare(req))
		panicIf(err)
		TypeInfoRes, err := serializeTypeInfo(prepare(response))
		panicIf(err)
		r := routeDef{
			ReqType:     req.String(),
			ResType:     response.String(),
			Method:      route.Method,
			Path:        route.Path,
			Handler:     route.Handler,
			FnInfo:      fnInfo,
			TypeInfoReq: string(TypeInfoReq),
			TypeInfoRes: string(TypeInfoRes),
		}

		if _, ok := m[fnInfo.PkgName]; !ok {
			m[fnInfo.PkgName] = make(map[string][]routeDef)
		}
		m[fnInfo.PkgName][fnInfo.StructName] = append(m[fnInfo.PkgName][fnInfo.StructName], r)
		normalizedPath := convertColonPathToBraces(r.Path)

		p := swagger.Paths.Find(normalizedPath)
		if p == nil {
			p = &spec.PathItem{}
			swagger.Paths[normalizedPath] = p
		}

		op := spec.NewOperation()

		parameters := []*spec.ParameterRef{}
		reqFields := reflect.VisibleFields(req)

		for _, field := range reqFields {
			if tag, ok := field.Tag.Lookup("query"); ok {
				parameters = append(parameters, &spec.ParameterRef{
					Value: &spec.Parameter{
						Name:     tag,
						In:       "query",
						Required: false, // TODO: document implementation with omitempty then resolve optional params here
						Schema:   spec.NewSchemaRef("", mapGoTypeToOpenAPISchema(field.Type)),
					},
				})
			} else if tag, ok := field.Tag.Lookup("url"); ok {
				parameters = append(parameters, &spec.ParameterRef{
					Value: &spec.Parameter{
						Name:     tag,
						In:       "path",
						Required: false, // TODO: document implementation with omitempty then resolve optional url parts here
						Schema:   spec.NewSchemaRef("", mapGoTypeToOpenAPISchema(field.Type)),
					},
				})
			}
		}

		op.Parameters = parameters

		if route.Method != "GET" {
			op.RequestBody = &spec.RequestBodyRef{
				Ref: typegen.RefReqPrefix + r.ReqType,
			}
		}

		if op.Responses == nil {
			op.Responses = spec.NewResponses()
		}
		resp := spec.NewResponse()
		description := "info"
		resp.Description = &description
		resp.Content = spec.NewContentWithSchemaRef(spec.NewSchemaRef(typegen.RefSchemaPrefix+r.ResType, nil), []string{"application/json"})
		op.AddResponse(200, resp)
		swagger.Components.RequestBodies[r.ReqType] = &spec.RequestBodyRef{
			Value: spec.NewRequestBody().WithContent(spec.NewContentWithSchemaRef(spec.NewSchemaRef(typegen.RefSchemaPrefix+r.ReqType, nil), []string{"application/json"})),
		}
		if p == nil {
			pi := &spec.PathItem{}
			swagger.Paths[r.Path] = pi
			p = pi
		}
		op.Tags = append(op.Tags, r.FnInfo.PkgName)
		p.SetOperation(r.Method, op)

	}

}

func convertColonPathToBraces(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) > 0 && part[0] == ':' {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func mapGoTypeToOpenAPISchema(t reflect.Type) *spec.Schema {
	switch t.Kind() {
	case reflect.Bool:
		return &spec.Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &spec.Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &spec.Schema{Type: "number"}
	case reflect.String:
		return &spec.Schema{Type: "string"}
	default:
		return &spec.Schema{Type: "string"}
	}
}
