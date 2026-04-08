package api2

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/starius/api2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtobuf(t *testing.T) {
	type ProtoRequest struct {
		Body *durationpb.Duration `use_as_body:"true" is_protobuf:"true"`
	}
	type ProtoResponse struct {
		Body *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
	}

	t1 := time.Date(2020, time.July, 10, 11, 30, 0, 0, time.UTC)

	protoHandler := func(ctx context.Context, req *ProtoRequest) (res *ProtoResponse, err error) {
		// Add passed duration to t1 and pass result back.
		duration := req.Body.AsDuration()
		t2 := t1.Add(duration)
		return &ProtoResponse{
			Body: timestamppb.New(t2),
		}, nil
	}

	routes := []api2.Route{
		{Method: http.MethodPost, Path: "/proto", Handler: protoHandler},
	}

	mux := http.NewServeMux()
	api2.BindRoutes(mux, routes)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := api2.NewClient(routes, server.URL)

	protoRes := &ProtoResponse{}
	err := client.Call(context.Background(), protoRes, &ProtoRequest{
		Body: durationpb.New(time.Second),
	})
	if err != nil {
		t.Errorf("request with protobuf failed: %v.", err)
	}
	got := protoRes.Body.AsTime().Format("2006-01-02 15:04:05")
	want := "2020-07-10 11:30:01"
	if got != want {
		t.Errorf("request with protobuf returned %q, want %q", got, want)
	}
}

func TestProtobufAcceptsBinaryBodyMislabeledAsJSON(t *testing.T) {
	type ProtoRequest struct {
		Body *durationpb.Duration `use_as_body:"true" is_protobuf:"true"`
	}
	type ProtoResponse struct {
		Body *timestamppb.Timestamp `use_as_body:"true" is_protobuf:"true"`
	}

	t1 := time.Date(2020, time.July, 10, 11, 30, 0, 0, time.UTC)

	protoHandler := func(ctx context.Context, req *ProtoRequest) (res *ProtoResponse, err error) {
		return &ProtoResponse{
			Body: timestamppb.New(t1.Add(req.Body.AsDuration())),
		}, nil
	}

	routes := []api2.Route{
		{Method: http.MethodPost, Path: "/proto", Handler: protoHandler},
	}

	mux := http.NewServeMux()
	api2.BindRoutes(mux, routes)
	server := httptest.NewServer(mux)
	defer server.Close()

	reqBody, err := proto.Marshal(durationpb.New(time.Second))
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, server.URL+"/proto", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http.DefaultClient.Do failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body=%q", resp.StatusCode, http.StatusOK, string(body))
	}
	if got := resp.Header.Get("Content-Type"); got != "application/x-protobuf" {
		t.Fatalf("content-type = %q, want application/x-protobuf", got)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll failed: %v", err)
	}

	got := &timestamppb.Timestamp{}
	if err := proto.Unmarshal(respBody, got); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	want := timestamppb.New(t1.Add(time.Second))
	if !proto.Equal(got, want) {
		t.Fatalf("response = %v, want %v", got, want)
	}
}
