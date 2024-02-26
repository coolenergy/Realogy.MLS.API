package main

import (
	"bitbucket.org/realogy_corp/mls-display-rules/internal/server"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"gopkg.in/urfave/cli.v1"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
)

var flags struct {
	Console bool
	ports   struct {
		service int
		grpc    int
		graphql int
	}
	host string
	env  string
}

func main() {
	app := cli.NewApp()
	app.Name = "MlsDisplayRulesService"
	app.Usage = "API for Mls Display Rules"
	flags := []cli.Flag{
		cli.BoolFlag{
			Name:        "console, cs",
			Usage:       "running in console mode",
			EnvVar:      "NO_AUZ",
			Destination: &flags.Console,
		},
		cli.IntFlag{
			Name:        "servicePort, sp",
			Usage:       "Service port to listen on (for endpoints)",
			EnvVar:      "PORT",
			Value:       8083,
			Destination: &flags.ports.service,
		},
		cli.IntFlag{
			Name:        "grpcPort, gp",
			Usage:       "Grpc port to listen on",
			EnvVar:      "GRPC_PORT",
			Value:       9981,
			Destination: &flags.ports.grpc,
		},
		cli.StringFlag{
			Name:        "Host",
			EnvVar:      "HOST",
			Value:       "localhost",
			Destination: &flags.host,
		},
		cli.StringFlag{
			Name:        "env",
			Value:       "local",
			Usage:       "env of deployment",
			EnvVar:      "ENV",
			Destination: &flags.env,
		},
	}
	app.Flags = flags
	app.Action = launch
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}

}

func launch(_ *cli.Context) error {

	// setup mongo connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	environment := server.Context{Environment: flags.env}
	mongodbPrefix := "mongodb+srv"

	awsConfig := &aws.Config{
		Region: aws.String("us-west-2"),
	}

	if environment.Environment == "local" {
		log.Printf("Using localstack for simulating AWS environment - ssm : [%v]", os.Getenv("AWS_LOCAL_ENDPOINT_SSM"))
		awsEndpointResolver := localAwsEndpointResolver("ssm", os.Getenv("AWS_LOCAL_ENDPOINT_SSM")) // for development testing.
		awsConfig.WithEndpointResolver(endpoints.ResolverFunc(awsEndpointResolver))

		mongodbPrefix = "mongodb"
	}

	var (
		databaseDetails = server.DatabaseDetails{}
		s               = session.Must(session.NewSession(awsConfig))
		ssmClient       = ssm.New(s)
	)

	err := server.LoadParametersCache(ssmClient, environment, &databaseDetails)
	if err != nil {
		log.Fatalf("Failed to load parameters : %v", err)
	}

	mongoConnectionString := mongodbPrefix+"://" + databaseDetails.MongoDbUser + ":" + databaseDetails.MongoDbPw + "@" + databaseDetails.MongoDbHost
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnectionString))

	defer client.Disconnect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB %v, error: %v", mongoConnectionString, err)
	}

	// start grpc server
	go server.StartService(server.StartServiceInput{
		GrpcPort:    flags.ports.grpc,
		MongoClient: client,
		Tracing:     false,
		Auth:        false,
	})

	log.Printf("GRPC Server listenning on port %v", flags.ports.grpc)

	// Start REST Gateway proxy
	h := server.NewGatewayProxy(ctx, server.GWInput{
		NoTracing:   true,
		ServicePort: flags.ports.service,
		GrpcPort:    flags.ports.grpc,
		Console:     flags.Console})

	log.Printf("REST API server listenning on port %v", flags.ports.service)

	addr := fmt.Sprintf(":%v", flags.ports.service)
	return http.ListenAndServe(addr, h)
}

func localAwsEndpointResolver(awsService string, awsEndpoint string) func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	return func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == awsService {
			return endpoints.ResolvedEndpoint{
				URL:           awsEndpoint,
				SigningRegion: region,
			}, nil
		}
		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}
}

func tracerWrap(handler http.Handler) http.Handler {
	return &ochttp.Handler{
		Handler: handler,
		StartOptions: trace.StartOptions{
			Sampler: trace.AlwaysSample(),
		},
	}
}

type loopbackTransport struct {
	handler http.Handler
}

func (l loopbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	l.handler.ServeHTTP(recorder, req)
	return recorder.Result(), nil
}
