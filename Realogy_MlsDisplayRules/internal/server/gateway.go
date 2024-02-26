package server

import (
	models "bitbucket.org/realogy_corp/mls-display-rules/internal/generated/realogy.com/api/mls/displayrules/v1"
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"log"
	"net/http"
	_ "time"
)

type GWInput struct {
	Console     bool
	ServicePort int
	GrpcPort    int
	NoTracing   bool
	Host        string
}

func NewGatewayProxy(ctx context.Context, flags GWInput) http.Handler {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	getMethod := fmt.Sprintf("%v:%v", flags.Host, flags.GrpcPort)
	err := models.RegisterMlsDisplayRulesServiceHandlerFromEndpoint(ctx, mux, getMethod, opts)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to register Agent Handler from endpoint \n"))
	}

	return mux
}
