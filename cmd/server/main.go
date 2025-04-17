package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/logger"
	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/repository"
	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/service"
	pb "github.com/pauloaugusto-dmf/tigerbeetle-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger.Init(false)
	// Parâmetros da linha de comando
	port := flag.Int("port", 50051, "Porta do servidor gRPC")
	clusterID := 0
	flag.Parse()

	// Divide os endereços do TigerBeetle
	addresses := "3000"

	// Inicializa o repositório TigerBeetle
	repo, err := repository.NewTigerBeetleRepository([]string{addresses}, uint64(clusterID))
	if err != nil {
		log.Fatalf("Falha ao inicializar repositório TigerBeetle: %v", err)
	}
	defer repo.Close()

	println("Conectado ao TigerBettle")

	// Inicializa o servidor gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Falha ao escutar na porta %d: %v", *port, err)
	}

	grpcServer := grpc.NewServer()

	// Registra o serviço financeiro
	financialService := service.NewFinancialService(repo)
	pb.RegisterFinancialServiceServer(grpcServer, financialService)

	// Habilita reflection para ferramentas como grpcurl
	reflection.Register(grpcServer)

	log.Printf("Servidor gRPC iniciado na porta %d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}

	log.Printf("Teste")
}
