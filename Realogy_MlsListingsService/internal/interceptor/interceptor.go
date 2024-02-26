package interceptor

import (
	"context"
	"log"
	"mlslisting/internal/config"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct {
	auth *config.Auth
}

func NewInterceptor(auth *config.Auth) *Interceptor {
	return &Interceptor{auth: auth}
}

//UnaryAuthInterceptor: for clients configured in config.yaml.
//If we do not get the token we just forward the request so that
//other users who are not configured in config.yml can bypass Authorization
func (i *Interceptor) UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("Interceptor Initiated")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("metadata is not provided.")
		return handler(ctx, req)
	}

	values := md["authorization"]
	if len(values) == 0 {
		log.Println("authorization token is not provided.")
		return handler(ctx, req)
	}
	bearerToken := values[0]
	accessToken := strings.Split(bearerToken, " ")[1]
	claims := jwt.MapClaims{}
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, claims)
	if err != nil {
		log.Println("Token cannot be parsed")
	}
	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Invalid Token")
	}
	cid := claim["cid"]
	//If ClientId and Endpoint do not match , it returns an error.
	if i.isPermissionDenied(cid.(string), info.FullMethod) {
		return nil, status.Errorf(codes.PermissionDenied, "Permission Denied")
	}

	m, err := handler(ctx, req)
	return m, err
}

func (i *Interceptor) isPermissionDenied(key, value string) bool {
	rules := strings.Split(i.auth.AccessRules, ";")
	accessRules := make(map[string][]string)
	for _, v := range rules {
		clientRules := strings.Split(v, ",")
		clientRules[1] = strings.ReplaceAll(clientRules[1], "[", "")
		clientRules[1] = strings.ReplaceAll(clientRules[1], "]", "")
		clientRules[1] = strings.ReplaceAll(clientRules[1], "\"", "")
		//Delimeter "*" is used to separate endpoints for respected clientID
		accessRules[clientRules[0]] = strings.Split(clientRules[1], "*")
	}

	//Considering key(cid) is not case sensitive.
	if v, ok := accessRules[strings.ToLower(key)]; ok {
		if !contains(value, v) {
			return true
		}
	}
	return false
}

func contains(search string, s []string) bool {
	for k := range s {
		if s[k] == search {
			return true
		}
	}
	return false

}
