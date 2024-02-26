//go:build integration

package service

import (
	models "bitbucket.org/realogy_corp/mls-display-rules/internal/generated/realogy.com/api/mls/displayrules/v1"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"os"
	"testing"
	"time"
)

var lis *bufconn.Listener

const (
	bufSize = 1024 * 1024
	port    = 9080
)

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	service := new(MlsDisplayRulesService)
	mongodbUrl := os.Getenv("MONGODB_URL")
	if mongodbUrl == "" {
		mongodbUrl = "localhost:27017"
	}
	service.MongoClient, _ = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:example@"+mongodbUrl))
	models.RegisterMlsDisplayRulesServiceServer(s, service)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func TestGetMlsDisplayRules(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := models.NewMlsDisplayRulesServiceClient(conn)
	request1 := new(models.GetMlsDisplayRulesByRequest)
	response1, e := client.GetMlsDisplayRules(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response1)
	assert.GreaterOrEqual(t, len(response1.MlsDisplayRules), 1)

	request2 := new(models.GetMlsDisplayRulesByRequest)
	request2.Offset = 0
	request2.Limit = 3
	response2, e := client.GetMlsDisplayRules(ctx, request2)
	assert.Nil(t, e)
	assert.NotNil(t, response2)
	assert.Equal(t, 3, len(response2.MlsDisplayRules))
}

func TestGetMlsDisplayRulesBySource(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	// send grpc request
	/*conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v", err)
	}
	defer conn.Close()*/

	client := models.NewMlsDisplayRulesServiceClient(conn)
	request1 := new(models.GetMlsDisplayRulesBySourceRequest)
	request1.SourceSystemKey = "NM_SWMLS"
	response1, e := client.GetMlsDisplayRulesBySource(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response1)
	assert.NotNil(t, response1.MlsDisplayRules)
	log.Println("response1: ", response1)
	log.Println("response1.MlsDisplayRules.LongName: ", response1.MlsDisplayRules.LongName)
	assert.Equal(t, "NM_SWMLS", response1.MlsDisplayRules.Source)
	assert.Equal(t, "Southwest Multiple Listing Service", response1.MlsDisplayRules.LongName)

	request2 := new(models.GetMlsDisplayRulesBySourceRequest)
	request2.SourceSystemKey = ""
	response2, e := client.GetMlsDisplayRulesBySource(ctx, request2)
	assert.NotNil(t, e)
	// Assertion takes first param as Expected and second param as actual. But for Contains(), its kind of the opposite
	assert.Contains(t, e.Error(), "could not process because input request is nil")
	assert.Nil(t, response2)

	request3 := new(models.GetMlsDisplayRulesBySourceRequest)
	request3.SourceSystemKey = "DoesNotExists"
	response3, e := client.GetMlsDisplayRulesBySource(ctx, request3)
	assert.NotNil(t, e)
	assert.Contains(t, e.Error(), "No documents found.")
	assert.Nil(t, response3)
}

func TestUpdateMlsDisplayRulesStatus(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := models.NewMlsDisplayRulesServiceClient(conn)
	request1 := new(models.UpdateMlsDisplayRulesStatusRequest)
	request1.MlsDisplayRulesStatus = new(models.MlsDisplayRulesStatus)
	request1.MlsDisplayRulesStatus.SourceSystemKey = "IA_CIBR"
	request1.MlsDisplayRulesStatus.IsActive = true
	response1, e := client.UpdateMlsDisplayRulesStatus(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response1)
	assert.NotNil(t, response1.MlsDisplayRules)
	assert.Equal(t, true, response1.MlsDisplayRules.IsActive)
	assert.Equal(t, "IA_CIBR", response1.MlsDisplayRules.Source)

	request1.MlsDisplayRulesStatus.IsActive = false
	response2, e := client.UpdateMlsDisplayRulesStatus(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response2)
	assert.NotNil(t, response2.MlsDisplayRules)
	assert.Equal(t, false, response2.MlsDisplayRules.IsActive)
	assert.Equal(t, "IA_CIBR", response2.MlsDisplayRules.Source)
}

func TestUpdateMlsDisplayRules(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := models.NewMlsDisplayRulesServiceClient(conn)
	request1 := new(models.UpdateMlsDisplayRulesDataRequest)
	request1.MlsDisplayRulesData = new(models.MlsDisplayRulesUpdateData)
	request1.MlsDisplayRulesData.SourceSystemKey = "MI_JMLS"
	request1.MlsDisplayRulesData.Disclaimer = "Test Disclaimer"
	request1.MlsDisplayRulesData.LongName = "Test Long Name"

	response1, e := client.UpdateMlsDisplayRulesData(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response1)
	assert.NotNil(t, response1.MlsDisplayRules)
	assert.Equal(t, "Test Disclaimer", response1.MlsDisplayRules.Disclaimer)
	assert.Equal(t, "Test Long Name", response1.MlsDisplayRules.LongName)
	assert.Equal(t, "MI_JMLS", response1.MlsDisplayRules.Source)

	request1.MlsDisplayRulesData.Disclaimer = "Test Disclaimer New"
	response2, e := client.UpdateMlsDisplayRulesData(ctx, request1)
	assert.Nil(t, e)
	assert.NotNil(t, response2)
	assert.NotNil(t, response2.MlsDisplayRules)
	assert.Equal(t, "Test Disclaimer New", response2.MlsDisplayRules.Disclaimer)
	assert.Equal(t, "Test Long Name", response2.MlsDisplayRules.LongName)
	assert.Equal(t, "MI_JMLS", response2.MlsDisplayRules.Source)
}

func TestStreamMlsDisplayRules(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := models.NewMlsDisplayRulesServiceClient(conn)
	resp, err := client.StreamMlsDisplayRules(ctx, new(empty.Empty))
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
