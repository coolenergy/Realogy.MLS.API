package service

import (
	models "bitbucket.org/realogy_corp/mls-display-rules/internal/generated/realogy.com/api/mls/displayrules/v1"
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var log = logrus.New()

type MlsDisplayRulesService struct {
	models.MlsDisplayRulesServiceServer
	MongoClient *mongo.Client
}

// Database and collection name from mongo
const mlsDb = "mls"
const displayRulesCollection = "display_rules"

//GetMlsDisplayRules returns all display rules. Apply limit-offset to get specific result
func (d *MlsDisplayRulesService) GetMlsDisplayRules(ctx context.Context, request *models.GetMlsDisplayRulesByRequest) (*models.GetMlsDisplayRulesByResponse, error) {

	if request == nil {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	response := &models.GetMlsDisplayRulesByResponse{}
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)

	filter := bson.D{{Key: "isActive", Value: true}}
	cur, err := collection.Find(ctx, filter, findOptions(request.Limit, request.Offset))
	defer cur.Close(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Errorf("No documents found : %v", err)
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			log.Errorf("Database internal error : %v", err)
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}

	for cur.Next(ctx) {
		var result models.MlsDisplayRules
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the document: %v , %s", err, result.Source)
		}
		response.MlsDisplayRules = append(response.MlsDisplayRules, &result)
	}
	return response, nil
}

// GetMlsDisplayRulesIgnoreStatus returns all display rules. Apply limit-offset to get specific result
func (d *MlsDisplayRulesService) GetMlsDisplayRulesIgnoreStatus(ctx context.Context, request *models.GetMlsDisplayRulesByRequest) (*models.GetMlsDisplayRulesByResponse, error) {

	if request == nil {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	response := &models.GetMlsDisplayRulesByResponse{}
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	filter := bson.D{}
	cur, err := collection.Find(ctx, filter, findOptions(request.Limit, request.Offset))
	defer cur.Close(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}

	for cur.Next(ctx) {
		var result models.MlsDisplayRules
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the document: %v , %s", err, result.Source)
		}
		response.MlsDisplayRules = append(response.MlsDisplayRules, &result)
	}
	return response, nil
}

//GetMlsDisplayRules returns all display rules. Apply limit-offset to get specific result
func (d *MlsDisplayRulesService) GetMlsDisplayRulesBySource(ctx context.Context, request *models.GetMlsDisplayRulesBySourceRequest) (*models.GetMlsDisplayRulesBySourceResponse, error) {

	if request == nil || request.SourceSystemKey == "" {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	response := &models.GetMlsDisplayRulesBySourceResponse{}
	var displayRule *models.MlsDisplayRules
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	filter := bson.D{{Key: "source", Value: request.SourceSystemKey},
		{Key: "isActive", Value: true}}
	cur := collection.FindOne(ctx, filter)
	err = cur.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}
	err = cur.Decode(&displayRule)
	if err != nil {
		log.Errorf("Unable to decode the document: %v", err)
	}
	response.MlsDisplayRules = displayRule
	return response, nil
}

//GetMlsDisplayRules returns all display rules. Apply limit-offset to get specific result
func (d *MlsDisplayRulesService) GetMlsDisplayRulesBySourceIgnoreStatus(ctx context.Context, request *models.GetMlsDisplayRulesBySourceRequest) (*models.GetMlsDisplayRulesBySourceResponse, error) {

	if request == nil || request.SourceSystemKey == "" {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	response := &models.GetMlsDisplayRulesBySourceResponse{}
	var displayRule *models.MlsDisplayRules
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	filter := bson.D{{Key: "source", Value: request.SourceSystemKey}}
	cur := collection.FindOne(ctx, filter)
	err = cur.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}
	err = cur.Decode(&displayRule)
	if err != nil {
		log.Errorf("Unable to decode the document: %v", err)
	}
	response.MlsDisplayRules = displayRule
	return response, nil
}

//StreamMlsDisplayRules streams on grpc all the display rules
func (d *MlsDisplayRulesService) StreamMlsDisplayRules(empty *empty.Empty, stream models.MlsDisplayRulesService_StreamMlsDisplayRulesServer) error {
	ctx := context.Background()
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Minute)
	cur, err := collection.Find(ctx, bson.D{}, findOptions)
	defer cur.Close(ctx)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return err
	}

	for cur.Next(ctx) {
		var result models.MlsDisplayRules
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the document: %v", err)
		}
		if err := stream.Send(&result); err != nil {
			log.Errorf("Error while streaming mls display rules: %v", err)
			return err
		}
	}
	stream.Context().Done()
	return nil
}

//StreamMlsDisplayRules streams on grpc all the display rules
func (d *MlsDisplayRulesService) StreamMlsDisplayRulesEvent(event *models.StreamMlsDisplayRulesEventRequest, stream models.MlsDisplayRulesService_StreamMlsDisplayRulesEventServer) error {
	ctx := context.Background()
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	var eventPipeline primitive.D
	if event.SourceSystemKey != "" {
		eventPipeline = append(eventPipeline, bson.E{Key: "source", Value: event.SourceSystemKey})
	}
	if event.EventType != "" {
		eventPipeline = append(eventPipeline, bson.E{Key: "operationType", Value: event.EventType})
	}
	if eventPipeline == nil {
		eventPipeline = bson.D{{}} // Subscribe to all the events
	}

	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{{"$or",
		bson.A{
			eventPipeline,
			bson.D{{"operationType", "update"}},
			bson.D{{"operationType", "insert"}},
			bson.D{{"operationType", "delete"}},
			bson.D{{"operationType", "replace"}},
		}}},
	}}}

	var changeStreamOptions options.ChangeStreamOptions
	changeStreamOptions = *options.ChangeStream().SetFullDocument(options.UpdateLookup)
	if event.ResumeEventId != "" {
		log.Printf("Resume mls display rules events after %s", event.ResumeEventId)
		changeStreamOptions.SetResumeAfter(bsonx.Doc{{"_data", bsonx.String(event.ResumeEventId)}})
	}
	cs, err := collection.Watch(ctx, pipeline, &changeStreamOptions)
	defer cs.Close(ctx)

	if err != nil {
		log.Errorf("Error while listening change streams ! %v", err)
		return stream.Context().Err()
	} else {
		log.Printf("Request has been made to listen for change : %s", event)
		for cs.Next(ctx) {
			var event DisplayRulesEvent
			err := cs.Decode(&event)
			if err != nil {
				log.Errorf("Unable to decode the mongo document: %v", err)
			}
			if &event != nil {
				eventMetaData := &models.EventMetaData{Data: event.EventData.EventId}
				result := &models.StreamMlsDisplayRulesEventResponse{MlsDisplayRules: event.DisplayRules, EventMetaData: eventMetaData, EventType: event.EventType}

				if err := stream.Send(result); err != nil {
					log.Errorf("Error while streaming mls display rules: %v", err)
					return err
				} else {
					// log.Printf("Sending mls %s event for %s",)
				}
			}
		}
	}
	stream.Context().Done()
	return nil
}

// UpdateMlsDisplayRulesStatus sends the updated response for which the active status was changed
func (d *MlsDisplayRulesService) UpdateMlsDisplayRulesStatus(ctx context.Context, request *models.UpdateMlsDisplayRulesStatusRequest) (*models.UpdateMlsDisplayRulesStatusResponse, error) {

	if request == nil || request.MlsDisplayRulesStatus == nil || request.MlsDisplayRulesStatus.SourceSystemKey == "" {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	condition := bson.D{{"source", request.MlsDisplayRulesStatus.SourceSystemKey}}
	updatedVal := bson.D{{"$set", bson.D{{"isActive", request.MlsDisplayRulesStatus.IsActive}}}}

	/**
	The update logic is to update the fields only when it has some data
	*/
	updatedSource, err := collection.UpdateOne(ctx, condition, updatedVal)
	if updatedSource.MatchedCount < 1 {
		log.Error("Source not found")
		return nil, errors.New("could not process because source is not found")
	}

	log.Info("Number of records updated : ", updatedSource.ModifiedCount)
	response := new(models.UpdateMlsDisplayRulesStatusResponse)
	filter := bson.D{{Key: "source", Value: request.MlsDisplayRulesStatus.SourceSystemKey},
		{Key: "isActive", Value: request.MlsDisplayRulesStatus.IsActive}}
	cur := collection.FindOne(ctx, filter)
	err = cur.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}
	var displayRule *models.MlsDisplayRules
	err = cur.Decode(&displayRule)
	if err != nil {
		log.Errorf("Unable to decode the document: %v", err)
	}
	response.MlsDisplayRules = displayRule
	return response, err
}

// UpdateMlsDisplayRulesStatus sends the updated response for which the active status was changed
func (d *MlsDisplayRulesService) UpdateMlsDisplayRulesData(ctx context.Context, request *models.UpdateMlsDisplayRulesDataRequest) (*models.UpdateMlsDisplayRulesDataResponse, error) {

	if request == nil || request.MlsDisplayRulesData == nil || request.MlsDisplayRulesData.SourceSystemKey == "" {
		log.Error("Input request was empty")
		return nil, errors.New("could not process because input request is nil")
	}
	var err error
	collection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	condition := bson.D{{"source", request.MlsDisplayRulesData.SourceSystemKey}}
	updateQuery := buildUpdateDisplayRulesBsonIgnoreNil(request)
	updatedSource, err := collection.UpdateOne(ctx, condition, updateQuery)
	if updatedSource.MatchedCount < 1 {
		log.Error("Source not found")
		return nil, errors.New("could not process because source is not found")
	}

	log.Info("Number of records updated : ", updatedSource.ModifiedCount)
	response := new(models.UpdateMlsDisplayRulesDataResponse)
	filter := bson.D{{Key: "source", Value: request.MlsDisplayRulesData.SourceSystemKey}}
	cur := collection.FindOne(ctx, filter)
	err = cur.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = status.Errorf(codes.NotFound, "No documents found.")
		} else {
			err = status.Errorf(codes.Internal, "Database internal error.")
		}
		return nil, err
	}
	var displayRule *models.MlsDisplayRules
	err = cur.Decode(&displayRule)
	if err != nil {
		log.Errorf("Unable to decode the document: %v", err)
	}
	response.MlsDisplayRules = displayRule
	return response, err
}

/**
Update display rules with empty
*/
func buildUpdateDisplayRulesBson(request *models.UpdateMlsDisplayRulesDataRequest) bson.D {

	if request.MlsDisplayRulesData == nil {
		return nil
	}

	var updatedBson bson.D
	updatedBson = append(updatedBson, bson.E{Key: "disclaimer", Value: request.MlsDisplayRulesData.Disclaimer})
	updatedBson = append(updatedBson, bson.E{Key: "longName", Value: request.MlsDisplayRulesData.LongName})
	updateQuery := bson.D{{"$set", updatedBson}}
	return updateQuery
}

/**
Update display rules ignore empty
*/
func buildUpdateDisplayRulesBsonIgnoreNil(request *models.UpdateMlsDisplayRulesDataRequest) bson.D {

	if request.MlsDisplayRulesData == nil {
		return nil
	}

	var updatedBson bson.D

	if request.MlsDisplayRulesData.Disclaimer != "" {
		updatedBson = append(updatedBson, bson.E{Key: "disclaimer", Value: request.MlsDisplayRulesData.Disclaimer})
	}
	if request.MlsDisplayRulesData.LongName != "" {
		updatedBson = append(updatedBson, bson.E{Key: "longName", Value: request.MlsDisplayRulesData.LongName})
	}

	updateQuery := bson.D{
		{"$set", updatedBson},
	}
	return updateQuery
}

func (d *MlsDisplayRulesService) HealthCheck(ctx context.Context, in *models.HealthRequest) (*models.HealthReply, error) {
	response := &models.HealthReply{Status: "Up"}

	mongoCollection := d.MongoClient.Database(mlsDb).Collection(displayRulesCollection)
	runCmdOpts := &options.RunCmdOptions{ReadPreference: readpref.Nearest(readpref.WithMaxStaleness(90 * time.Second))}
	err := mongoCollection.Database().RunCommand(ctx, bsonx.Doc{{"ping", bsonx.String("1")}}, runCmdOpts).Decode(&response)
	if err != nil {
		log.Errorf("Error while pinging mongodb for health check : %v", err)
		return nil, err
	}
	return response, nil
}

func findOptions(limit int32, offset int32) *options.FindOptions {
	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Minute) // TODO: Configurable ?
	if limit == 0 {
		findOptions.SetLimit(20) // Default value is "20"
	} else {
		findOptions.SetLimit(int64(limit))
	}
	findOptions.SetSkip(int64(offset))
	return findOptions
}

type DisplayRulesEvent struct {
	EventData        EventData               `json:"_id" bson:"_id"`
	EventType        string                  `json:"operationType" bson:"operationType"`
	EventTime        primitive.Timestamp     `json:"clusterTime" bson:"clusterTime"`
	DisplayRules     *models.MlsDisplayRules `json:"fullDocument" bson:"fullDocument"`
	EventDocumentKey *EventDocumentKey       `json:"documentKey" bson:"documentKey"`
}

type EventData struct {
	EventId string `bson:"_data"`
}

type EventDocumentKey struct {
	Id string `bson:"_id"`
}
