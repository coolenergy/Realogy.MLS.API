package gateway

import (
	"context"
	"fmt"
	"mlslisting/internal/config"
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	log "github.com/sirupsen/logrus"

	"net"

	"google.golang.org/grpc"
)

const (
	health = "health"
)

// Endpoint describes a gRPC endpoint
type Endpoint struct {
	Network, Addr string
}

// Options is a set of options to be passed to Run
type Options struct {
	// Addr is the address to listen
	Addr string

	// GRPCServer defines an endpoint of a gRPC service
	GRPCServer Endpoint

	// Mux is a list of options to be passed to the server-gateway multiplexer
	Mux []gwruntime.ServeMuxOption
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func Run(ctx context.Context, opts Options) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := dial(ctx, opts.GRPCServer.Network, opts.GRPCServer.Addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			log.Errorf("Failed to close a client connection to the gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/"+health, healthServer(conn))

	gw, err := newGateway(ctx, conn)
	if err != nil {
		return err
	}
	mux.Handle("/", gw)

	s := &http.Server{
		Addr:    opts.Addr,
		Handler: allowCORS(mux),
	}
	go func() {
		<-ctx.Done()
		log.Infof("Shutting down http gateway server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http gateway server: %v", err)
		}
	}()

	log.Infof("Grpc gateway listening at %s", opts.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Grpc gateway failed to listen and serve: %v", err)
		return err
	}

	return nil
}

func WaitForGateway(ctx context.Context, port uint16) error {
	ch := time.After(10 * time.Second)

	var err error
	for {
		if r, err := http.Get(fmt.Sprintf("http://localhost:%d/%s", port, health)); err == nil {
			if r.StatusCode == http.StatusOK {
				return nil
			}
			err = fmt.Errorf("grpc localhost:%d returned an unexpected status %d", port, r.StatusCode)
		}

		log.Infof("Waiting for localhost:%d to get ready", port)
		select {
		case <-ctx.Done():
			return err
		case <-ch:
			return err
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func Start(ctx context.Context, config config.Config) error {

	if err := Run(ctx, Options{
		Addr: fmt.Sprintf(":%d", config.Gateway.Port),
		GRPCServer: Endpoint{
			Network: config.Gateway.Network,
			Addr:    fmt.Sprintf("localhost:%d", config.Grpc.Port),
		},
	}); err != nil {
		return err
	}

	// wait for gateway
	if err := WaitForGateway(ctx, config.Gateway.Port); err != nil {
		log.Errorf("Failed to verify the health of Gateway server due to: %v", err)
	}

	return nil
}

// Creates new server gateway server which translates HTTP into gRPC request
func newGateway(ctx context.Context, conn *grpc.ClientConn) (http.Handler, error) {

	mux := gwruntime.NewServeMux(
		gwruntime.WithMarshalerOption(gwruntime.MIMEWildcard, &gwruntime.HTTPBodyMarshaler{
			Marshaler: &gwruntime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   false,
					EmitUnpopulated: true,
					Multiline:       false,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		gwruntime.WithIncomingHeaderMatcher(httpHeaderMatcher),
		gwruntime.WithStreamErrorHandler(streamErrorHandler),
	)

	for _, f := range []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error{
		pb.RegisterMlsListingServiceHandler,
	} {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}

	return mux, nil
}

// interrupt stream error from grpc and wrap it with http error.
func streamErrorHandler(ctx context.Context, err error) *status.Status {
	code := codes.Internal
	msg := "unexpected error"
	if s, ok := status.FromError(err); ok {
		code = s.Code()
		log.Debugf("streaming error code: %v", code)
		msg = s.Message()
		log.Debugf("streaming error message: %s", msg)
	}

	return status.Newf(code, msg)
}

func dial(ctx context.Context, network, addr string) (*grpc.ClientConn, error) {
	switch network {
	case "tcp":
		return dialTCP(ctx, addr)
	case "unix":
		return dialUnix(ctx, addr)
	default:
		return nil, fmt.Errorf("unsupported network type %q", network)
	}
}

// dialTCP creates a client connection via TCP. "addr" must be a valid TCP address with a port number.
func dialTCP(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, grpc.WithInsecure())
}

// dialUnix creates a client connection via a unix domain socket. "addr" must be a valid path to the socket.
func dialUnix(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	d := func(addr string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", addr, timeout)
	}
	return grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithDialer(d))
}

// match headers in the http request and restrict the one that should be forwarded to grpc context.
func httpHeaderMatcher(key string) (string, bool) {
	switch key { // keys are converted to TitleCase.
	case "Apikey": // apiKey is transformed to "Apikey".
		return key, true
	case "Authorization", "authorization": // authorization is transformed to "Authorization".
		return key, true
	default: // expand this to allow more headers. restricted only to "apiKey" for now.
		return key, false
	}
}
