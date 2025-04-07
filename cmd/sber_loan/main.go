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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/Dorji/sberInterview/api/protos/services"
	"github.com/Dorji/sberInterview/internal/loanservice"
	"github.com/Dorji/sberInterview/internal/loanservice/interceptors"
	loadconfig "github.com/Dorji/sberInterview/internal/loanservice/load_config"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
)

var (
	httpAddr string
	grpcAddr string
)

func main() {
	// Load configuration
	config, err := loadconfig.LoadConfig("config.yml")
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
			interceptors.RecoveryUnaryInterceptor,
			interceptors.LoggingUnaryInterceptor,
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
		Handler: interceptors.RecoveryMiddleware(interceptors.LoggingMiddleware(httpMux)),
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
