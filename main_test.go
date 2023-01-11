package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"buf.build/gen/go/coobeet/coobeet/bufbuild/connect-go/coobeet/v1/coobeetv1connect"
	coobeetv1 "buf.build/gen/go/coobeet/coobeet/protocolbuffers/go/coobeet/v1"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoServer(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.Handle(coobeetv1connect.NewEchoServiceHandler(NewEchoServer()))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	connectClient := coobeetv1connect.NewEchoServiceClient(
		server.Client(),
		server.URL,
	)
	grpcClient := coobeetv1connect.NewEchoServiceClient(
		server.Client(),
		server.URL,
		connect.WithGRPC(),
	)
	clients := []coobeetv1connect.EchoServiceClient{connectClient, grpcClient}

	t.Run("Echo", func(t *testing.T) {
		for _, client := range clients {
			res, err := client.Echo(context.Background(), connect.NewRequest(&coobeetv1.EchoRequest{
				Message: "hello",
			}))
			require.Nil(t, err)
			assert.Equal(t, "hello", res.Msg.Message)
		}
	})
}
