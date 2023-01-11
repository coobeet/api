package main

import (
	"context"
	"net/http"

	"buf.build/gen/go/coobeet/coobeet/bufbuild/connect-go/coobeet/v1/coobeetv1connect"
	coobeetv1 "buf.build/gen/go/coobeet/coobeet/protocolbuffers/go/coobeet/v1"
	"github.com/bufbuild/connect-go"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type echoServer struct {
	coobeetv1connect.UnimplementedEchoServiceHandler
}

func (s *echoServer) Echo(
	ctx context.Context,
	req *connect.Request[coobeetv1.EchoRequest],
) (*connect.Response[coobeetv1.EchoResponse], error) {
	return connect.NewResponse(&coobeetv1.EchoResponse{
		Message: req.Msg.Message,
	}), nil
}

func main() {
	mux := http.NewServeMux()
	mux.Handle(coobeetv1connect.NewEchoServiceHandler(&echoServer{}))
	handler := cors.AllowAll().Handler(mux)
	handler = h2c.NewHandler(handler, &http2.Server{})
	http.ListenAndServe("localhost:8080", handler)
}
