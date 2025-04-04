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

	"github.com/Dorji/sberInterview/api/protos/services"
	"github.com/Dorji/sberInterview/internal/loanservice"
	"github.com/Dorji/sberInterview/internal/loanservice/storage"
)

const (
	httpAddr = ":8080"
	grpcAddr = ":50051"
)

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
            statusCode:     http.StatusOK, // Значение по умолчанию
        }

        next.ServeHTTP(recorder, r)

        duration := time.Since(start).Nanoseconds()
        log.Printf(
            "%s status_code: %d, duration: %d ns",
            time.Now().Format("2006/01/02 15:04:05"),
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
		time.Now().Format("2006/01/02 15:04:05"),
		statusCode,
		duration,
	)

	return resp, err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%s gRPC panic in %s: %v", 
				time.Now().Format("2006/01/02 15:04:05"),
				info.FullMethod, 
				r)
			err = status.Errorf(codes.Internal, "Internal Server Error")
		}
	}()
	return handler(ctx, req)
}

// HTTP сервер
type httpServer struct{}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	default:
		http.NotFound(w, r)
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
	// Регистрируем gRPC Gateway endpoints
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Создаем gRPC сервер
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor,
			loggingUnaryInterceptor,
		),
	)
	registerGRPCHandlers(grpcSrv)

	// 2. Запускаем gRPC сервер
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

	// 3. Создаем HTTP роутер
	httpMux := http.NewServeMux()

	// 4. Добавляем обычные HTTP обработчики
	httpMux.Handle("/", &httpServer{})

	// 5. Создаем gRPC Gateway роутер
	gwMux := runtime.NewServeMux()
	if err := registerHTTPHandlers(ctx, gwMux); err != nil {
		log.Fatalf("failed to register HTTP handlers: %v", err)
	}

	// 6. Объединяем роутеры
	httpMux.Handle("/v1/", gwMux) // Все gRPC HTTP-ручки будут доступны по /v1/

	// 7. Настраиваем HTTP сервер
	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: recoveryMiddleware(loggingMiddleware(httpMux)),
	}

	// 8. Graceful shutdown
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

	// 9. Запускаем HTTP сервер
	log.Printf("HTTP server listening on %s", httpAddr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}