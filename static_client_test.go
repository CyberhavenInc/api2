package api2

import (
	"context"
	"testing"
)

func TestGetMethod(t *testing.T) {
	type FooRequest struct{}
	type FooResponse struct{}
	var foo func(ctx context.Context, req *FooRequest) (*FooResponse, error)

	type BarArgs struct{}
	type BarReply struct{}
	var bar func(ctx context.Context, req *BarArgs) (*BarReply, error)

	type BazRequest struct{}
	type BazResponse struct{}
	type BazService interface {
		Baz(ctx context.Context, req *BazRequest) (*BazResponse, error)
	}
	var baz BazService

	type GetRequest struct{}
	type GetResponse struct{}
	var getHandler func(ctx context.Context, req *GetRequest) (*GetResponse, error)

	type GetWithBodyRequest struct {
		Param string
	}
	type GetWithBodyResponse struct {
		Result string
	}
	var getWithBodyHandler func(ctx context.Context, req *GetWithBodyRequest) (*GetWithBodyResponse, error)

	var badFoo func(ctx context.Context, req *FooRequest) (*BarReply, error)
	var badFoo2 func(ctx context.Context, req []int) ([]string, error)

	testCases := []struct {
		name     string
		handler  interface{}
		request  string
		response string
		wantErr  bool
	}{
		{
			name:     "Foo",
			handler:  foo,
			request:  "FooRequest",
			response: "FooResponse",
			wantErr:  false,
		},
		{
			name:     "Bar",
			handler:  bar,
			request:  "BarArgs",
			response: "BarReply",
			wantErr:  false,
		},
		{
			name:     "Baz",
			handler:  Method(&baz, "Baz"),
			request:  "BazRequest",
			response: "BazResponse",
			wantErr:  false,
		},
		{
			name:    "BadFoo",
			handler: badFoo,
			wantErr: true,
		},
		{
			name:    "BadFoo2",
			handler: badFoo2,
			wantErr: true,
		},
		{
			name:     "Get",
			handler:  getHandler,
			request:  "GetRequest",
			response: "GetResponse",
			wantErr:  false,
		},
		{
			name:     "GetWithBody",
			handler:  getWithBodyHandler,
			request:  "GetWithBodyRequest",
			response: "GetWithBodyResponse",
			wantErr:  false,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			name, request, response, err := getMethod(tc.handler)
			if tc.wantErr {
				if err == nil {
					t.Errorf("case %d: want error, not got", i)
				}
			} else {
				if err != nil {
					t.Errorf("case %d: got error: %v", i, err)
				}
				if name != tc.name {
					t.Errorf("case %d: want name %q, got %q", i, tc.name, name)
				}
				if request != tc.request {
					t.Errorf("case %d: want request %q, got %q", i, tc.request, request)
				}
				if response != tc.response {
					t.Errorf("case %d: want response %q, got %q", i, tc.response, response)
				}
			}
		})
	}
}
