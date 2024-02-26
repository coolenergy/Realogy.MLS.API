package server

import (
	models "bitbucket.org/realogy_corp/mls-display-rules/internal/generated/realogy.com/api/mls/displayrules/v1"
	"bitbucket.org/realogy_corp/mls-display-rules/internal/service"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

type StartServiceInput struct {
	GrpcPort    int
	MongoClient *mongo.Client
	Tracing     bool
	Auth        bool
	OPAUrl      string
}

var opaURL string

var (
	// Metrics registry
	PromRegistry = prometheus.NewRegistry()

	// Grpc grpc metrics
	PrometheusGrpcMetrics = prom.NewServerMetrics()
)

/**/
func StartService(in StartServiceInput) {

	opaURL = in.OPAUrl

	s := grpc.NewServer()

	service := &service.MlsDisplayRulesService{}
	service.MongoClient = in.MongoClient

	s = grpc.NewServer(grpc.StreamInterceptor(PrometheusGrpcMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(PrometheusGrpcMetrics.UnaryServerInterceptor()),
	)
	models.RegisterMlsDisplayRulesServiceServer(s, service)
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(in.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
