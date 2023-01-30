package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buf.build/gen/go/coobeet/coobeet/bufbuild/connect-go/coobeet/v1/coobeetv1connect"
	coobeetv1 "buf.build/gen/go/coobeet/coobeet/protocolbuffers/go/coobeet/v1"
	"github.com/bufbuild/connect-go"
	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type EchoServer struct {
	coobeetv1connect.UnimplementedEchoServiceHandler
}

func NewEchoServer() *EchoServer {
	return &EchoServer{}
}

func (s *EchoServer) Echo(
	ctx context.Context,
	req *connect.Request[coobeetv1.EchoRequest],
) (*connect.Response[coobeetv1.EchoResponse], error) {
	return connect.NewResponse(&coobeetv1.EchoResponse{
		Message: req.Msg.Message,
	}), nil
}

func createChannel() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	return stopCh, func() {
		close(stopCh)
	}
}

func start(server *http.Server) {
	log.Println("application started")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	} else {
		log.Println("application stopped gracefully")
	}
}

func shutdown(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	} else {
		log.Println("application shutdowned")
	}
}

func newCORS() *cors.Cors {
	// To let web developers play with the demo service from browsers, we need a
	// very permissive CORS setup.
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins, which effectively disables CORS.
			return true
		},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{
			// Content-Type is in the default safelist.
			"Accept",
			"Accept-Encoding",
			"Accept-Post",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Content-Encoding",
			"Grpc-Accept-Encoding",
			"Grpc-Encoding",
			"Grpc-Message",
			"Grpc-Status",
			"Grpc-Status-Details-Bin",
		},
		// Let browsers cache CORS information for longer, which reduces the number
		// of preflight requests. Any changes to ExposedHeaders won't take effect
		// until the cached data expires. FF caps this value at 24h, and modern
		// Chrome caps it at 2h.
		MaxAge: int(2 * time.Hour / time.Second),
	})
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	compress1KB := connect.WithCompressMinBytes(1024)
	mux.Handle(coobeetv1connect.NewEchoServiceHandler(NewEchoServer(), compress1KB))
	mux.Handle(grpchealth.NewHandler(
		grpchealth.NewStaticChecker(coobeetv1connect.EchoServiceName),
		compress1KB,
	))
	mux.Handle(grpcreflect.NewHandlerV1(
		grpcreflect.NewStaticReflector(coobeetv1connect.EchoServiceName),
		compress1KB,
	))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(
		grpcreflect.NewStaticReflector(coobeetv1connect.EchoServiceName),
		compress1KB,
	))

	addr := "localhost:8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	s := &http.Server{
		Addr:              addr,
		Handler:           h2c.NewHandler(newCORS().Handler(mux), &http2.Server{}),
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		MaxHeaderBytes:    8 * 1024, // 8KiB
	}
	go start(s)

	stopCh, closeCh := createChannel()
	defer closeCh()
	log.Println("notified:", <-stopCh)

	shutdown(context.Background(), s)
}
