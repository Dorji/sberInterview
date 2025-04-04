package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"

	"github.com/Dorji/sberInterview/api/protos/services"
	"github.com/Dorji/sberInterview/internal/loanservice"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
)

// Config represents application configuration
type Config struct {
	HTTP struct {
		Port string `yaml:"port"`
	} `yaml:"http"`
	GRPC struct {
		Port string `yaml:"port"`
	} `yaml:"grpc"`
}

func loadConfig(path string) (*Config, error) {
	config := &Config{
		HTTP: struct{ Port string }{Port: "8080"},
		GRPC: struct{ Port string }{Port: "50051"},
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %v, using defaults", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return config, fmt.Errorf("error parsing config file: %v, using defaults", err)
	}

	return config, nil
}

// HTTP Middleware
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start).Nanoseconds()
		log.Printf(
			"%s %s status_code: %d, duration: %d ns",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			duration,
		)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("HTTP panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// gRPC Interceptors
func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	statusCode := status.Code(err)
	duration := time.Since(start).Nanoseconds()
	log.Printf(
		"%s status_code: %d, duration: %d ns",
		info.FullMethod,
		statusCode,
		duration,
	)

	return resp, err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("gRPC panic in %s: %v", info.FullMethod, r)
			err = status.Errorf(codes.Internal, "Internal Server Error")
		}
	}()
	return handler(ctx, req)
}

func registerGRPCHandlers(grpcSrv *grpc.Server) {
	myCache := storage.NewLoanCache()
	ls, err := loanservice.NewLoanService(myCache)
	if err != nil {
		log.Fatalf("start NewLoanService error: %v", err)
	}
	services.RegisterLoanServiceServer(grpcSrv, ls)
	reflection.Register(grpcSrv)
}

func registerHTTPHandlers(ctx context.Context, mux *runtime.ServeMux) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := services.RegisterLoanServiceHandlerFromEndpoint(
		ctx,
		mux,
		grpcAddr,
		opts,
	); err != nil {
		return fmt.Errorf("failed to register gRPC gateway: %v", err)
	}

	return nil
}

var (
	httpAddr string
	grpcAddr string
)

func main() {
	// Load configuration
	config, err := loadConfig("config.yml")
	if err != nil {
		log.Printf("Config warning: %v", err)
	}

	httpAddr = ":" + config.HTTP.Port
	grpcAddr = ":" + config.GRPC.Port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Create gRPC server
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor,
			loggingUnaryInterceptor,
		),
	)
	registerGRPCHandlers(grpcSrv)

	// 2. Start gRPC server
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// 3. Create HTTP router
	httpMux := http.NewServeMux()
	
	// 4. Create gRPC Gateway router
	gwMux := runtime.NewServeMux()
	if err := registerHTTPHandlers(ctx, gwMux); err != nil {
		log.Fatalf("failed to register HTTP handlers: %v", err)
	}

	// 5. Combine routers
	httpMux.Handle("/", gwMux)

	// 6. Configure HTTP server
	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: recoveryMiddleware(loggingMiddleware(httpMux)),
	}

	// 7. Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Println("Shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}

		grpcSrv.GracefulStop()
		log.Println("Servers stopped gracefully")
	}()

	// 8. Start HTTP server
	log.Printf("HTTP server listening on %s", httpAddr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}