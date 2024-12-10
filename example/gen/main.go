package main

import (
	"github.com/CyberhavenInc/api2"
	"github.com/CyberhavenInc/api2/example"
)

func main() {
	api2.GenerateClient(example.GetRoutes)
	api2.GenerateTSClient(&api2.TypesGenConfig{
		OutDir:    "./ts-types",
		Blacklist: []api2.BlacklistItem{{Service: "Hello"}},
		Routes:    []interface{}{example.GetRoutes},
		Types: []interface{}{
			&example.CustomType{},
			&example.CustomType2{},
		},
		EnumsWithPrefix: true,
	})
	api2.GenerateOpenApiSpec(&api2.TypesGenConfig{
		OutDir: "./openapi",
		Routes: []interface{}{example.GetRoutes},
		Types: []interface{}{
			&example.EchoRequest{},
			&example.CustomType2{},
		},
	})
	api2.GenerateYamlClient(&api2.YamlTypesGenConfig{
		OutDir: "./rego",
		Routes: []interface{}{example.GetRoutes},
	})

}
