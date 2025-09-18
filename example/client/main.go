package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/starius/api2/example"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	client, err := example.NewClient("http://127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	helloRes, err := client.Hello(ctx, &example.HelloRequest{
		Key: "secret password",
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Echo(ctx, &example.EchoRequest{
		Session: helloRes.Session,
		Text:    "test",
	})
	if err == nil {
		panic("expected an error")
	}

	echoRes, err := client.Echo(ctx, &example.EchoRequest{
		Session: helloRes.Session,
		User:    "good-user",
		Text:    "test",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(echoRes.Text)

	sinceRes, err := client.Since(ctx, &example.SinceRequest{
		Session: helloRes.Session,
		Body:    timestamppb.New(time.Date(2020, time.July, 10, 11, 30, 0, 0, time.UTC)),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sinceRes.Body.AsDuration())

	streamRes, err := client.Stream(ctx, &example.StreamRequest{
		Session: helloRes.Session,
		Body:    io.NopCloser(strings.NewReader("abc xyz")),
	})
	if err != nil {
		panic(err)
	}
	streamResBytes, err := io.ReadAll(streamRes.Body)
	if err != nil {
		panic(err)
	}
	if err := streamRes.Body.Close(); err != nil {
		panic(err)
	}
	fmt.Println(string(streamResBytes))

	redirectRes, err := client.Redirect(ctx, &example.RedirectRequest{
		ID: "user123",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(redirectRes.Status, redirectRes.URL)

	rawRes, err := client.Raw(ctx, &example.RawRequest{
		Token: []byte("secret"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(rawRes.Token))

	advancedRes, err := client.AdvancedWildcard(ctx, &example.AdvancedWildcardRequest{
		ParamA: "A",
		ParamB: "B/B",
		ParamC: "C",
		ParamD: "D/D/D",
		ParamE: "E",
		ParamF: "F/F/F/F",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(advancedRes.Token))
	if "AdvancedWildcard: A B/B C D/D/D E F/F/F/F" != string(advancedRes.Token) {
		panic("unexpected advanced wildcard response")
	}

	basicRes, err := client.BasicWildcard(ctx, &example.BasicWildcardRequest{
		ParamA: "A",
		ParamB: "B/B/B",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(basicRes.Token))
	if "BasicWildcard: A B/B/B" != string(basicRes.Token) {
		panic("unexpected basic wildcard response: " + string(basicRes.Token))
	}
}
