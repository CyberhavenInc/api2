package api2

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	"go.uber.org/goleak"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestValidateRequestResponse(t *testing.T) {
	cases := []struct {
		obj       interface{}
		request   bool
		path      string
		wantPanic bool
	}{
		{
			obj: struct {
				Foo string `json:"foo"`
			}{},
			request: true,
		},
		{
			obj: struct {
				Foo string `json:"foo"`
			}{},
			request: false,
		},
		{
			obj: struct {
				Foo string `query:"foo"`
			}{},
			request: true,
		},
		{
			obj: struct {
				Foo string `query:"foo"`
			}{},
			request:   false,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `header:"foo"`
			}{},
			request: true,
		},
		{
			obj: struct {
				Foo string `header:"foo"`
			}{},
			request: false,
		},
		{
			obj: struct {
				Foo string `cookie:"foo"`
			}{},
			request: true,
		},
		{
			obj: struct {
				Foo string `cookie:"foo"`
			}{},
			request:   false,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo http.Cookie `cookie:"foo"`
			}{},
			request: false,
		},
		{
			obj: struct {
				Status int `use_as_status:"true"`
			}{},
			request:   false,
			wantPanic: false,
		},
		{
			obj: struct {
				Status string `use_as_status:"true"`
			}{},
			request:   false,
			wantPanic: true,
		},
		{
			obj: struct {
				Status int `use_as_status:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},

		{
			obj: struct {
				Foo int `json:"foo" use_as_status:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo int `use_as_body:"true" use_as_status:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo int `json:"foo" use_as_status:"true"`
			}{},
			request:   false,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo int `use_as_body:"true" use_as_status:"true"`
			}{},
			request:   false,
			wantPanic: true,
		},
		{
			obj: struct {
				Status  int `use_as_status:"true"`
				Status2 int `use_as_status:"true"`
			}{},
			request:   false,
			wantPanic: true,
		},

		{
			obj: struct {
				Body struct {
					Foo int `json:"foo"`
				} `use_as_body:"true"`
				Bar int `json:"bar"`
			}{},
			request:   true,
			wantPanic: true,
		},

		{
			obj: struct {
				Foo string `json:"foo" query:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `header:"foo" query:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo" header:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo" cookie:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `header:"foo" cookie:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `cookie:"foo" query:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo" header:"foo" query:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo" header:"foo" query:"foo" cookie:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},

		{
			obj: struct {
				Foo string `url:"foo"`
			}{},
			request:   true,
			path:      "/post/:foo",
			wantPanic: false,
		},
		{
			obj: struct {
				Foo string `url:"foo"`
			}{},
			request:   true,
			path:      "",
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo"`
			}{},
			request:   true,
			path:      "/post/:foo",
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `json:"foo" url:"foo"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `url:"foo"`
			}{},
			request:   false,
			wantPanic: true,
		},

		{
			obj: struct {
				Foo *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo []byte `use_as_body:"true" is_raw:"true"`
			}{},
			request:   true,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo string `use_as_body:"true" is_raw:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo []byte `use_as_body:"true" is_raw:"true"`
			}{},
			request:   false,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true" is_raw:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo  *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
				Foo2 *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   false,
			wantPanic: false,
		},
		{
			obj: struct {
				AnyProtobuf proto.Message `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: false,
		},
		{
			obj: struct {
				AnyProtobuf proto.Message `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   false,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo string `use_as_body:"true" is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo *timestamppb.Timestamp `is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},

		{
			obj: struct {
				Foo io.ReadCloser `use_as_body:"true" is_stream:"true"`
			}{},
			request:   true,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo io.ReadCloser `use_as_body:"true" is_stream:"true"`
			}{},
			request:   false,
			wantPanic: false,
		},
		{
			obj: struct {
				Foo *io.ReadCloser `use_as_body:"true" is_stream:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo string `use_as_body:"true" is_stream:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
		{
			obj: struct {
				Foo io.ReadCloser `use_as_body:"true" is_stream:"true" is_protobuf:"true"`
			}{},
			request:   true,
			wantPanic: true,
		},
	}

	for i, tc := range cases {
		var message interface{}
		gotPanic := func() (gotPanic bool) {
			defer func() {
				if r := recover(); r != nil {
					gotPanic = true
					message = r
				}
			}()
			validateRequestResponse(reflect.TypeOf(tc.obj), tc.request, tc.path)
			return
		}()
		if gotPanic != tc.wantPanic {
			t.Errorf("case %d: gotPanic=%v, wantPanic=%v, message=%v", i, gotPanic, tc.wantPanic, message)
		}
	}
}

type HelloRequest struct {
}

type HelloResponse struct {
}

type ServiceStruct struct {
}

func (s *ServiceStruct) Hello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{}, nil
}

type ServiceInterface interface {
	Hello(ctx context.Context, req *HelloRequest) (*HelloResponse, error)
}

func TestMethod(t *testing.T) {
	var servicePtrNil *ServiceStruct
	var serviceInterfaceNil ServiceInterface

	servicePtr := &ServiceStruct{}
	serviceInterface := ServiceInterface(&ServiceStruct{})

	cases := []struct {
		method                                   interface{}
		pkgFull, pkgName, structName, methodName string
	}{
		{
			method:     Method(&servicePtr, "Hello"),
			pkgFull:    "github.com/CyberhavenInc/api2",
			pkgName:    "api2",
			structName: "ServiceStruct",
			methodName: "Hello",
		},
		{
			method:     Method(&servicePtrNil, "Hello"),
			pkgFull:    "github.com/CyberhavenInc/api2",
			pkgName:    "api2",
			structName: "ServiceStruct",
			methodName: "Hello",
		},
		{
			method:     Method(&serviceInterface, "Hello"),
			pkgFull:    "github.com/CyberhavenInc/api2",
			pkgName:    "api2",
			structName: "ServiceInterface",
			methodName: "Hello",
		},
		{
			method:     Method(&serviceInterfaceNil, "Hello"),
			pkgFull:    "github.com/CyberhavenInc/api2",
			pkgName:    "api2",
			structName: "ServiceInterface",
			methodName: "Hello",
		},
	}

	for i, tc := range cases {
		method := tc.method.(*interfaceMethod)

		pkgFull, pkgName, structName, methodName := method.FuncInfo()
		if pkgFull != tc.pkgFull {
			t.Errorf("case %d: for pkgFull = %q, want %q", i, pkgFull, tc.pkgFull)
		}
		if pkgName != tc.pkgName {
			t.Errorf("case %d: for pkgName = %q, want %q", i, pkgName, tc.pkgName)
		}
		if structName != tc.structName {
			t.Errorf("case %d: for structName = %q, want %q", i, structName, tc.structName)
		}
		if methodName != tc.methodName {
			t.Errorf("case %d: for method = %q, want %q", i, methodName, tc.methodName)
		}

		f := method.Func()
		path := "/hello"
		validateHandler(reflect.TypeOf(f), path)
	}
}
