package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/repository"
	"github.com/pauloaugusto-dmf/tigerbeetle-service/internal/service"
	pb "github.com/pauloaugusto-dmf/tigerbeetle-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Parâmetros da linha de comando
	port := flag.Int("port", 50051, "Porta do servidor gRPC")
	tbAddresses := flag.String("tb-addresses", "localhost:3000", "Endereços do TigerBeetle separados por vírgula")
	clusterID := flag.Uint("cluster-id", 0, "ID do cluster TigerBeetle")
	flag.Parse()

	// Divide os endereços do TigerBeetle
	addresses := strings.Split(*tbAddresses, ",")

	// Inicializa o repositório TigerBeetle
	repo, err := repository.NewTigerBeetleRepository(addresses, uint64(*clusterID))
	if err != nil {
		log.Fatalf("Falha ao inicializar repositório TigerBeetle: %v", err)
	}
	defer repo.Close()

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
}
