package server

import (
	"context"
	"fmt"
	"mlslisting/internal/interceptor"
	"mlslisting/internal/services"
	"reflect"

	"contrib.go.opencensus.io/exporter/aws"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"

	"mlslisting/internal/config"
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"net"
	"net/http"
	"time"

	codecs "github.com/amsokol/mongo-go-driver-protobuf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Config           *config.Config
	MongoClient      *mongo.Client
	MongoDatabase    *mongo.Database
	MongoCollections map[string]string
}

// creates mongodb connection, prometheus server and server server.
func (s *Server) Start(ctx context.Context) error {
	// mongodb
	s.initMongo(ctx)

	// start prometheus server
	s.startPrometheus()

	// grpc server
	if err := s.Run(ctx, s.Config.Grpc.Network, fmt.Sprintf(":%d", s.Config.Grpc.Port)); err != nil {
		return err
	}

	return nil
}

// mongodb initialization
func (s *Server) initMongo(ctx context.Context) {
	s.MongoClient = createMongoClient(ctx, &s.Config.MongoDB)
	s.MongoDatabase = s.MongoClient.Database(s.Config.MongoDB.Name)
	s.MongoCollections = s.Config.MongoDB.Collections
}

// mongodb client initialization
func createMongoClient(ctx context.Context, mongoConfig *config.MongoDBConfig) *mongo.Client {
	uri := config.GenerateMongoUrl(mongoConfig)

	// add more types as needed to handle nil (for some reason "omitempty" tag is being ignored).
	types := []interface{}{
		"",
		false,
		0.0,
		int32(0),
		int64(0),
		uint(0),
		uint32(0),
		uint64(0),
		float32(0),
		float64(0),
		time.Time{},
	}

	registry := codecs.Register(bson.NewRegistryBuilder())
	for _, v := range types {
		t := reflect.TypeOf(v)
		defDecoder, err := bson.DefaultRegistry.LookupDecoder(t)
		if err != nil {
			log.Panicf("Unable to use mongodb registry for decoding: %s", err)
		}
		registry.RegisterDecoder(t, &customDecoder{defDecoder, reflect.Zero(t)})
	}

	clientOptions := options.Client().ApplyURI(uri).
		SetRegistry(registry.Build()).
		SetConnectTimeout(time.Minute).
		SetServerSelectionTimeout(time.Minute).
		SetReadPreference(readpref.Nearest()).
		SetAuth(options.Credential{
			Username: mongoConfig.User,
			Password: mongoConfig.Pass,
		})

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	go func() {
		defer mongoClient.Disconnect(ctx)
		<-ctx.Done()
	}()

	if err != nil {
		log.Fatalf("Unable to connect to mongodb: %s ", err)
	}

	err = mongoClient.Ping(ctx, readpref.Nearest(readpref.WithMaxStaleness(90*time.Second)))
	if err != nil {
		log.Fatalf("Unable to verify the connection with mongodb: %s", err)
	} else {
		log.Info("Connected to mongodb !")
	}
	return mongoClient
}

// Prometheus Server
func (s *Server) startPrometheus() {
	promHTTPServer := &http.Server{Handler: promhttp.HandlerFor(config.PromRegistry, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("0.0.0.0:%d", s.Config.Prometheus.Port)}

	go func() {
		if err := promHTTPServer.ListenAndServe(); err != nil {
			log.Errorf("Unable to start a prometheus http server")
		}
	}()
}

// cleanup the connections
func (s *Server) Cleanup(ctx context.Context) {
	go func() {
		defer s.MongoClient.Disconnect(ctx)
	}()
}

// Run Grpc Server
func (s *Server) Run(ctx context.Context, network, address string) error {

	l, err := net.Listen(network, address)
	if err != nil {
		return err
	}

	defer func() {
		if err := l.Close(); err != nil {
			log.Errorf("Failed to close %s %s: %v", network, address, err)
		}
	}()

	// tracing and grpc metrics
	if s.Config.Tracing {
		exporter, err := aws.NewExporter()
		if err != nil {
			log.Error("Failed to create the AWS X-Ray exporter in gRPC server", zap.Error(err))
		}
		defer exporter.Close()
		trace.RegisterExporter(exporter)
		if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
			log.Error("Failed to register gRPC server views", zap.Error(err))
		}
		if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
			log.Error("Failed to register gRPC client views", zap.Error(err))
		}
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}
	ip := interceptor.NewInterceptor(&s.Config.Api.Auth)
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
		grpc.StreamInterceptor(config.PrometheusGrpcMetrics.StreamServerInterceptor()),
		grpc.ChainUnaryInterceptor(config.PrometheusGrpcMetrics.UnaryServerInterceptor(), ip.UnaryAuthInterceptor),
	)

	reflection.Register(grpcServer) // enable server reflection

	// initialize grpc metrics
	config.PrometheusGrpcMetrics.InitializeMetrics(grpcServer)

	// register
	pb.RegisterMlsListingServiceServer(grpcServer, &services.Service{MongoDatabase: s.MongoDatabase,
		ListingsCollection: s.MongoCollections["listings"],
		MaxQueryTimeSecs:   s.Config.MongoDB.MaxQueryTimeSecs,
		Pagination:         &s.Config.Api.Pagination,
		Stream:             &s.Config.Api.Stream,
		BySource:           &s.Config.Api.BySource,
		ByAddress:          &s.Config.Api.ByAddress})

	go func() {
		defer grpcServer.GracefulStop()
		<-ctx.Done()
	}()

	return grpcServer.Serve(l)
}
