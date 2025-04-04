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

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gen/go/protos/services" 
)

// HTTP сервер
type httpServer struct {
	// здесь могут быть зависимости вашего HTTP сервера
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Реализуйте ваши HTTP обработчики здесь
	switch r.URL.Path {
	case "/health":
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	default:
		http.NotFound(w, r)
	}
}

// gRPC сервер
type grpcServer struct {
	services.UnimplementedLoanServiceServer // замените YourService на ваш реальный сервис
}

// Реализуйте методы gRPC сервера здесь
func (s *grpcServer) YourMethod(ctx context.Context, req *proto.LoanRequest) (*services.LoanResponse, error) {
	// Реализация метода
	return &services.LoanResponse{Result: "Hello, " + req.Name}, nil
}

func main() {
	// Создаем канал для graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Настройка HTTP сервера
	httpAddr := ":8080"
	httpSrv := &http.Server{
		Addr:    httpAddr,
		Handler: &httpServer{},
	}

	// Запуск HTTP сервера
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Настройка gRPC сервера
	grpcAddr := ":50051"
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	services.RegisterLoanServiceServer(grpcSrv, &grpcServer{})
	reflection.Register(grpcSrv) // для удобства разработки (можно убрать в production)

	// Запуск gRPC сервера
	go func() {
		log.Printf("gRPC server listening on %s", grpcAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	<-done
	log.Println("Server stopped by signal")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	
	grpcSrv.GracefulStop()
	log.Println("Server exited properly")
}