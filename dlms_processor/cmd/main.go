package main

import (
	"dlmsprocessor/api"
	"dlmsprocessor/proto"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

const (
	GRPC_PORT = ":50051"
	HTTP_PORT = ":7070"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Start gRPC server in goroutine
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", GRPC_PORT)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		proto.RegisterDLMSProcessorServer(grpcServer, api.NewDLMSProcessorAPI())
		fmt.Printf("gRPC server is running on port %s\n", GRPC_PORT)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server in goroutine
	go func() {
		defer wg.Done()
		// Create a separate gRPC server for grpc-web
		grpcServer := grpc.NewServer()
		proto.RegisterDLMSProcessorServer(grpcServer, api.NewDLMSProcessorAPI())
		wrappedGrpc := grpcweb.WrapServer(grpcServer)

		// HTTP router (fallback to static UI if not gRPC)
		httpHandler := func(w http.ResponseWriter, r *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
				EnableCORS(wrappedGrpc).ServeHTTP(w, r)
				return
			}
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", httpHandler)
		fmt.Printf("HTTP server is running on port %s\n", HTTP_PORT)
		if err := http.ListenAndServe(HTTP_PORT, mux); err != nil {
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	// Wait for both servers
	wg.Wait()
}

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set necessary headers for CORS
		w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust in production
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-Agent, X-Grpc-Web")

		// Check for preflight request
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
