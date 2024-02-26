package services

import (
	"context"
	"mlslisting/internal/config"
	"mlslisting/internal/mlsvalidation"
	"mlslisting/internal/transformer"

	"github.com/aws/aws-sdk-go/aws"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/golang/protobuf/ptypes"

	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"strings"
	"time"

	"github.com/chidiwilliams/flatbson"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.MlsListingServiceServer
	MongoDatabase      *mongo.Database
	ListingsCollection string
	MaxQueryTimeSecs   int
	Pagination         *config.PaginationConfig
	Stream             *config.StreamConfig
	BySource           *config.BySource
	ByAddress          *config.ByAddress
}

type MlsEventResponse struct {
	EventData        EventData           `json:"_id" bson:"_id"`
	EventType        *string             `json:"operationType" bson:"operationType"`
	EventTime        primitive.Timestamp `json:"clusterTime" bson:"clusterTime"`
	MlsListing       *pb.MlsListing      `json:"fullDocument" bson:"fullDocument"`
	EventDocumentKey *EventDocumentKey   `json:"documentKey" bson:"documentKey"`
}

type EventData struct {
	EventId *string `bson:"_data"`
}

type FullDocument struct {
}

type EventDocumentKey struct {
	Id *string `bson:"_id"`
}

func (s *Service) GetMlsListingByListingId(ctx context.Context, in *pb.GetMlsListingByListingIdRequest) (*pb.GetMlsListingByListingIdResponse, error) {

	ctx, span := trace.StartSpan(ctx, "/listingById")
	defer span.End()

	err := validation.Errors{
		"ListingId": validation.Validate(in.ListingId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingByListingIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "listing_id", Value: bson.D{{"$in", bson.A{in.ListingId, strings.ReplaceAll(in.ListingId, "-", "_")}}}})
	if in.SourceSystemKey != "" {
		pipeline = append(pipeline, bson.E{Key: "source_system_key", Value: in.SourceSystemKey})
	}
	if in.PostalCode != "" {
		pipeline = append(pipeline, bson.E{Key: "property.location.address.postal_code", Value: bson.M{"$eq": in.PostalCode}})
	}
	// mongodb find options
	findOptions := options.Find()
	findOptions.SetCollation(&options.Collation{Locale: "en", Strength: 2}) // case insensitive search
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := mongoCollection.Find(ctx, &pipeline, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}
	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// incase of error, log and process next item TODO: Metrics for failed items
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in.ListingId)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) UpdateMlsListingByListingId(ctx context.Context, in *pb.UpdateMlsListingByListingIdRequest) (*pb.UpdateMlsListingByListingIdResponse, error) {

	ctx, span := trace.StartSpan(ctx, "/updatelistingById")
	defer span.End()
	log.Printf("updating mls listing with listingID: %s\n and SourceSystemKey %s\n", in.ListingId, in.SourceSystemKey)
	err := validation.Errors{
		"ListingId":       validation.Validate(in.ListingId, validation.Required),
		"SourceSystemKey": validation.Validate(in.SourceSystemKey, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var filter primitive.D
	filter = append(filter, bson.E{Key: "listing_id", Value: in.ListingId})
	filter = append(filter, bson.E{Key: "source_system_key", Value: in.SourceSystemKey})

	listingFromDB := &pb.MlsListing{}
	if err := mongoCollection.FindOne(context.TODO(), filter, &options.FindOneOptions{}).Decode(listingFromDB); err != nil {
		if err == mongo.ErrNoDocuments {
			msg := fmt.Sprintf("Unable to find Realogy Listing for given listingID %s", in.ListingId)
			log.Errorf(msg)
			return nil, status.Errorf(codes.NotFound, msg)
		} else {
			msg := "unable to update listing."
			log.Errorf("%v:ListingDecode Error: %v", msg, err)
			return nil, status.Errorf(codes.Internal, msg)
		}
	}

	// validate business rules for updating a listing
	err = mlsvalidation.ValidateUpdateListing(in, listingFromDB)
	if err != nil {
		return nil, err
	}

	// transform to UpdateMlsListing
	updatedListing := transformer.TransformUpdateListingInputToUpdateMlsListing(in)

	if err != nil {
		msg := fmt.Sprintf("Unable to Update Realogy Listing for given listingID %s", in.ListingId)
		log.Errorf("%v:Copier Error: %v", msg, err)
		return nil, status.Errorf(codes.Internal, msg)
	}

	flattenedBson, flatErr := flatbson.Flatten(&updatedListing)
	if flatErr != nil {
		msg := fmt.Sprintf("Unable to update Realogy Listing for given listingID %s", in.ListingId)
		log.Errorf("%v:Flattenbson Error: %v", msg, flatErr)
		return nil, status.Errorf(codes.Internal, msg)
	}

	// if the user only input 0/nil values to update or no valid updates at all then the last change date will be the only update in the bson. We
	// have the last change date in 2 places which is why we use 2 here.
	if len(flattenedBson) == 2 {
		log.Errorf("Validation Error. No Fields to update.")
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Unable to update empty values."))
	}

	update := bson.M{
		"$set": flattenedBson,
	}
	after := options.After
	findOneAndUpdateOptions := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	err = mongoCollection.FindOneAndUpdate(ctx, filter, update, &findOneAndUpdateOptions).Decode(&listingFromDB)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			msg := fmt.Sprintf("Unable to find Realogy Listing for given listingID %s", in.ListingId)
			log.Errorf("no documents matched to update %v", err)
			return nil, status.Error(codes.NotFound, msg)
		}
		msg := fmt.Sprintf("error updating doc %v", in.ListingId)
		return nil, status.Error(codes.Internal, msg)
	}
	resp := pb.UpdateMlsListingByListingIdResponse{
		MlsListings: listingFromDB,
	}

	return &resp, nil
}

func (s *Service) AddMlsListings(ctx context.Context, in *pb.MlsListingInput) (*pb.AddListingsResponse, error) {

	var insertedBy string
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil {
		apiKey := md.Get("apikey")
		if len(apiKey) != 0 {
			insertedBy = apiKey[0]
		}
		log.Printf("request has been made by %s to insert mls listing", md.Get("apikey"))
	}

	// validations for request attributes.
	err := validation.Errors{
		"ListingId":          validation.Validate(in.ListingId, validation.Required),
		"RdmSourceSystemKey": validation.Validate(in.RdmSourceSystemKey, validation.Required, validation.In("SOLO", "ELL", "LC")),
		"StandardStatus":     validation.Validate(in.Property.Listing.StandardStatus, validation.Required, validation.In("ACTIVE", "INACTIVE", "SOLD", "CANCELED", "HOLD", "UNKNOWN", "EXPIRED", "TEMP", "TERMINATED", "PENDING", "WITHDRAWN")),
		"PropertyType":       validation.Validate(in.Property.PropertyType, validation.Required, validation.In("SFR", "MFD", "CONDO", "COOP", "TOWNHOUSE", "MFR", "LAND", "FARM", "RENTAL", "COMMERCIAL_SALE", "COMMERCIAL_LEASE")),
		"PriceInput":         validation.Validate(in.Property.Listing.Price, validation.Required),
	}.Filter()
	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}
	in.Property.Listing.ListingId = in.ListingId

	// validate business rules for a new listing
	err = mlsvalidation.ValidateInsertListing(in)
	if err != nil {
		return nil, err
	}

	mlsListing := transformer.TransformListingInputToMlsListing(in)
	doc := bson.D{
		{Key: "_id", Value: in.RdmSourceSystemKey + "_" + in.ListingId},
		{Key: "listing_id", Value: in.ListingId},
		{Key: "rdm_source_system_Key", Value: in.RdmSourceSystemKey},
		{Key: "property", Value: mlsListing.Property},
		{Key: "last_changed_date", Value: time.Now()},
		// Restrict the POST endpoint to only ELL, SOLO & LC sources. Since these are
		// internal sources, for internal sources, the rdm_source_system_key and source_system_key is same
		{Key: "source_system_key", Value: in.RdmSourceSystemKey},
		{Key: "inserted_by", Value: insertedBy},
	}

	var d pb.MlsListing

	// mongodb.Collection.InsertOne
	result, err := s.MongoDatabase.Collection(s.ListingsCollection).InsertOne(context.TODO(), doc)

	update := bson.M{
		"$set": result,
	}
	after := options.After
	findOneAndUpdateOptions := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	if err := s.MongoDatabase.Collection(s.ListingsCollection).FindOneAndUpdate(context.TODO(), bson.D{{Key: "_id", Value: in.RdmSourceSystemKey + "_" + in.ListingId}}, update, findOneAndUpdateOptions).Decode(&d); err != nil {
		msg := fmt.Sprintf("error while inserting document , Listing with %v already exists in the database", in.ListingId)
		return nil, status.Error(codes.AlreadyExists, msg)
	}

	return &pb.AddListingsResponse{MlsListings: &d}, nil
}

func (s *Service) GetMlsListingByListingGuid(ctx context.Context, in *pb.GetMlsListingByListingGuidRequest) (*pb.GetMlsListingByListingGuidResponse, error) {
	err := validation.Errors{
		"ListingGuid": validation.Validate(in.ListingGuid, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingByListingGuidResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var filter primitive.M
	if in.SourceSystemKey != "" {
		filter = bson.M{"dash.listing_guid": in.ListingGuid, "source_system_key": in.SourceSystemKey}
	} else {
		filter = bson.M{"dash.listing_guid": in.ListingGuid}
	}

	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := mongoCollection.Find(ctx, &filter, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}
	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// incase of error, log and process next item TODO: Metrics for failed items
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in.ListingGuid)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByCity(ctx context.Context, in *pb.GetMlsListingsByCityRequest) (*pb.GetMlsListingsByCityResponse, error) {

	err := validation.Errors{
		"City": validation.Validate(in.City, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByCityResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	if in.State != "" {
		pipeline = append(pipeline, bson.E{Key: "property.location.address.city", Value: in.City}, bson.E{Key: "property.location.address.state_or_province", Value: in.State})
	} else {
		pipeline = append(pipeline, bson.E{Key: "property.location.address.city", Value: in.City})
	}
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	findOptions := s.findOptions(in.Limit, in.Offset)
	findOptions.SetCollation(&options.Collation{Locale: "en", Strength: 2}) // index with collation should exists.
	cur, err := mongoCollection.Find(ctx, &pipeline, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByState(ctx context.Context, in *pb.GetMlsListingsByStateRequest) (*pb.GetMlsListingsByStateResponse, error) {
	err := validation.Errors{
		"State": validation.Validate(in.State, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByStateResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.location.address.state_or_province", Value: in.State})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByPostalCode(ctx context.Context, in *pb.GetMlsListingsByPostalCodeRequest) (*pb.GetMlsListingsByPostalCodeResponse, error) {
	err := validation.Errors{
		"PostalCode": validation.Validate(in.PostalCode, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByPostalCodeResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.location.address.postal_code", Value: in.PostalCode})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. // TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in.PostalCode)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in.PostalCode))
	}
	return response, nil
}

func (s *Service) GetMlsListingBySource(ctx context.Context, in *pb.GetMlsListingsBySourceRequest) (*pb.GetMlsListingsBySourceResponse, error) {
	err := validation.Errors{
		"SourceSystemKey": validation.Validate(in.SourceSystemKey, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsBySourceResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	//Comma separated source system key
	sanitizedSourceSystemKey := strings.ReplaceAll(in.SourceSystemKey, " ", "")
	if sanitizedSourceSystemKey != "" {
		sourceSlice := strings.Split(sanitizedSourceSystemKey, ",")
		if len(sourceSlice) > 0 {
			pipeline = append(pipeline, bson.E{Key: "source_system_key", Value: bson.M{"$in": sourceSlice}})
		}
	}
	//Comma separated ListingAgentGuid
	sanitizedListingAgentGuid := strings.ReplaceAll(in.ListingAgentGuid, " ", "")
	if sanitizedListingAgentGuid != "" {
		agentGuidSlice := strings.Split(sanitizedListingAgentGuid, ",")
		if len(agentGuidSlice) > 0 {
			pipeline = append(pipeline, bson.E{Key: "dash.listing_agent_guid", Value: bson.M{"$in": agentGuidSlice}})
		}
	}

	// filter listings from a given date/time.
	if in.LastChangeTimestamp != nil {

		ts := timestamppb.Timestamp{Seconds: in.LastChangeTimestamp.Seconds, Nanos: 0}
		lastChangeTs := ts.AsTime()

		allowedLastChangeTs := time.Now().AddDate(0, 0, -s.BySource.AllowedLastChangeDays).UTC()
		if lastChangeTs.Unix() > allowedLastChangeTs.Unix() {
			log.Infof("Received request to get mls listings by source for start time: %v", lastChangeTs)
			pipeline = append(pipeline, bson.E{Key: "last_change_date", Value: bson.M{"$gte": lastChangeTs}})
		} else {
			log.Errorf("Listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays)
			return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays))
		}
	}
	if in.ListAgentMasterId != "" {
		pipeline = append(pipeline, bson.E{Key: "master_id.list_agent_master_id", Value: in.ListAgentMasterId})
	}
	if in.ListOfficeMasterId != "" {
		pipeline = append(pipeline, bson.E{Key: "master_id.list_office_master_id", Value: in.ListOfficeMasterId})
	}
	if in.CompanyMasterId != "" {
		pipeline = append(pipeline, bson.E{Key: "master_id.company_master_id", Value: in.CompanyMasterId})
	}

	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in.SourceSystemKey)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByAgentId(ctx context.Context, in *pb.GetMlsListingsByAgentIdRequest) (*pb.GetMlsListingsByAgentIdResponse, error) {
	err := validation.Errors{
		"ListAgentMlsId": validation.Validate(in.ListAgentMlsId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByAgentIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.listing.agent_office.list_agent.list_agent_mls_id", Value: in.ListAgentMlsId})
	if in.SourceSystemKey != "" {
		pipeline = append(pipeline, bson.E{Key: "source_system_key", Value: in.SourceSystemKey})
	}
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByAgentMasterId(ctx context.Context, in *pb.GetMlsListingsByAgentMasterIdRequest) (*pb.GetMlsListingsByAgentMasterIdResponse, error) {
	err := validation.Errors{
		"ListAgentMasterId": validation.Validate(in.ListAgentMasterId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByAgentMasterIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "master_id.list_agent_master_id", Value: in.ListAgentMasterId})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	findOptions := s.findOptions(in.Limit, in.Offset)
	findOptions.SetCollation(&options.Collation{Locale: "en", Strength: 2})
	cur, err := mongoCollection.Find(ctx, &pipeline, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByAgentGuid(ctx context.Context, in *pb.GetMlsListingsByAgentGuidRequest) (*pb.GetMlsListingsByAgentGuidResponse, error) {
	err := validation.Errors{
		"ListingAgentGuid": validation.Validate(in.ListingAgentGuid, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByAgentGuidResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "dash.listing_agent_guid", Value: in.ListingAgentGuid})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByAddress(ctx context.Context, in *pb.GetMlsListingsByAddressRequest) (*pb.GetMlsListingsByAddressResponse, error) {

	err := validation.Errors{
		"UnparsedAddress": validation.Validate(strings.TrimSpace(in.UnparsedAddress), validation.Required, validation.NilOrNotEmpty, validation.Length(10, 0)),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	log.Debugf("Listings for address: %s", in)
	response := &pb.GetMlsListingsByAddressResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	searchQuery := "property.location.address.unparsed_address:\"" + in.UnparsedAddress + "\""

	if in.City != "" {
		searchQuery = searchQuery + " AND property.location.address.city:\"" + in.City + "\""
	}
	if in.State != "" {
		searchQuery = searchQuery + " AND property.location.address.state_or_province:\"" + in.State + "\""
	}
	if in.PostalCode != "" {
		searchQuery = searchQuery + " AND property.location.address.postal_code:\"" + in.PostalCode + "\""
	}

	searchPipeline := bson.A{
		bson.D{{"$search",
			bson.D{
				bson.E{Key: "index", Value: s.ByAddress.SearchIndex},
				bson.E{Key: "queryString",
					Value: bson.D{
						{"defaultPath", "property.location.address.unparsed_address"},
						{"query", searchQuery},
					},
				}},
		}},
	}

	opts := options.Aggregate()
	cur, err := mongoCollection.Aggregate(ctx, searchPipeline, opts)

	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsBySubdivision(ctx context.Context, in *pb.GetMlsListingsBySubdivisionRequest) (*pb.GetMlsListingsBySubdivisionResponse, error) {
	err := validation.Errors{
		"SubdivisionName": validation.Validate(in.SubdivisionName, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsBySubdivisionResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.location.area.subdivision_name", Value: in.SubdivisionName})

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByOfficeMasterId(ctx context.Context, in *pb.GetMlsListingsByOfficeMasterIdRequest) (*pb.GetMlsListingsByOfficeMasterIdResponse, error) {
	err := validation.Errors{
		"ListOfficeMasterId": validation.Validate(in.ListOfficeMasterId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByOfficeMasterIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "master_id.list_office_master_id", Value: in.ListOfficeMasterId})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	findOptions := s.findOptions(in.Limit, in.Offset)
	findOptions.SetCollation(&options.Collation{Locale: "en", Strength: 2})
	cur, err := mongoCollection.Find(ctx, &pipeline, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByCompanyMasterId(ctx context.Context, in *pb.GetMlsListingsByCompanyMasterIdRequest) (*pb.GetMlsListingsByCompanyMasterIdResponse, error) {
	err := validation.Errors{
		"CompanyMasterId": validation.Validate(in.CompanyMasterId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByCompanyMasterIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "master_id.company_master_id", Value: in.CompanyMasterId})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	findOptions := s.findOptions(in.Limit, in.Offset)
	findOptions.SetCollation(&options.Collation{Locale: "en", Strength: 2})
	cur, err := mongoCollection.Find(ctx, &pipeline, findOptions)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByCompanyStaffId(ctx context.Context, in *pb.GetMlsListingsByCompanyStaffIdRequest) (*pb.GetMlsListingsByCompanyStaffIdResponse, error) {
	err := validation.Errors{
		"CompanyStaffMasterId": validation.Validate(in.CompanyStaffMasterId, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByCompanyStaffIdResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "master_id.company_staff_master_id", Value: in.CompanyStaffMasterId})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

func (s *Service) GetMlsListingsByCompanyStaffGuid(ctx context.Context, in *pb.GetMlsListingsByCompanyStaffGuidRequest) (*pb.GetMlsListingsByCompanyStaffGuidResponse, error) {
	err := validation.Errors{
		"CompanyStaffGuid": validation.Validate(in.CompanyStaffGuid, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.GetMlsListingsByCompanyStaffGuidResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "dash.company_staff_guid", Value: in.CompanyStaffGuid})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls listings for %s", in)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for %s", in))
	}
	return response, nil
}

// Sold Listings API accepts "startDate" and "endDate" to get sold listings. The "endDate" is optional and defaults to 6 months from the "startDate"
func (s *Service) GetMlsSoldListings(ctx context.Context, in *pb.GetMlsSoldListingsRequest) (*pb.GetMlsSoldListingsResponse, error) {

	response := &pb.GetMlsSoldListingsResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var pipeline primitive.D
	pipeline, err := SoldListingsPipeline(in, pipeline)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid date range to find sold listings [%s]. %v", in, err.Error()))
	}

	cur, err := mongoCollection.Find(ctx, &pipeline, s.findOptions(in.Limit, in.Offset))
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching mls %s", in))
	}

	// iterate mongo cursor and create response
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			// log error and process next item. no need to abort the loop. TODO: Metrics
			log.Errorf("Unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		log.Errorf("Unable to find mls sold listings for the given date range")
		return nil, status.Errorf(codes.NotFound, fmt.Sprint("Unable to find mls sold listings for the given date range"))
	}
	return response, nil
}

func (s *Service) StreamMlsListingByCity(in *pb.GetMlsListingsByCityRequest, stream pb.MlsListingService_StreamMlsListingByCityServer) error {
	err := validation.Errors{
		"City": validation.Validate(in.City, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return err //TODO: Fix me
	}

	collection := s.MongoDatabase.Collection(s.ListingsCollection)
	ctx, _ := context.WithCancel(context.Background())

	var pipeline primitive.D
	if in.State != "" {
		pipeline = append(pipeline, bson.E{Key: "property.location.address.city", Value: in.City}, bson.E{Key: "property.location.address.state_or_province", Value: in.State})
	} else {
		pipeline = append(pipeline, bson.E{Key: "property.location.address.city", Value: in.City})
	}
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := collection.Find(ctx, pipeline, findOptions)
	defer cur.Close(ctx)
	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
	}
	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the mongo document: %v", err)
		}
		if err := stream.Send(&result); err != nil {
			log.Errorf("Error while streaming mls listings: %v", err)
			return err
		}
	}
	stream.Context().Done()
	return nil
}

func (s *Service) StreamMlsListingByState(in *pb.GetMlsListingsByStateRequest, stream pb.MlsListingService_StreamMlsListingByStateServer) error {
	err := validation.Errors{
		"State": validation.Validate(in.State, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return err //TODO: Fix me
	}

	collection := s.MongoDatabase.Collection(s.ListingsCollection)
	ctx, _ := context.WithCancel(context.Background())

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.location.address.state_or_province", Value: in.State})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := collection.Find(ctx, pipeline, findOptions)
	defer cur.Close(ctx)

	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
	}

	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the mongo document: %v", err)
		}
		if err := stream.Send(&result); err != nil {
			log.Errorf("Error while streaming mls listings: %v", err)
			return err
		}
	}
	stream.Context().Done()
	return nil
}

func (s *Service) StreamMlsListingByPostalCode(in *pb.GetMlsListingsByPostalCodeRequest, stream pb.MlsListingService_StreamMlsListingByPostalCodeServer) error {
	err := validation.Errors{
		"PostalCode": validation.Validate(in.PostalCode, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return err //TODO: Fix me
	}

	collection := s.MongoDatabase.Collection(s.ListingsCollection)
	ctx, _ := context.WithCancel(context.Background())

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "property.location.address.postal_code", Value: in.PostalCode})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := collection.Find(ctx, pipeline, findOptions)
	defer cur.Close(ctx)

	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
	}

	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the mongo document: %v", err)
		}
		if err := stream.Send(&result); err != nil {
			log.Errorf("Error while streaming mls listings: %v", err)
			return err
		}
	}
	stream.Context().Done()
	return nil
}

func (s *Service) StreamMlsListingBySource(in *pb.GetMlsListingsBySourceRequest, stream pb.MlsListingService_StreamMlsListingBySourceServer) error {
	err := validation.Errors{
		"SourceSystemKey": validation.Validate(in.SourceSystemKey, validation.Required),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return err //TODO: Fix me
	}

	collection := s.MongoDatabase.Collection(s.ListingsCollection)
	ctx, _ := context.WithCancel(context.Background())

	var pipeline primitive.D
	pipeline = append(pipeline, bson.E{Key: "source_system_key", Value: in.SourceSystemKey})
	pipeline = append(pipeline, aggregatePipelineFilter(in.Filter)...)

	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	cur, err := collection.Find(ctx, pipeline, findOptions)
	defer cur.Close(ctx)

	if err != nil {
		log.Errorf("Error while processing the request to search mls: %v", err)
	}

	for cur.Next(ctx) {
		var result pb.MlsListing
		err := cur.Decode(&result)
		if err != nil {
			log.Errorf("Unable to decode the mongo document: %v", err)
		}
		if err := stream.Send(&result); err != nil {
			log.Errorf("Error while streaming mls listings: %v", err)
			return err
		}
	}
	stream.Context().Done()
	return nil
}

/*
 API to stream mls events. Clients can either subscribe to events specific to source system key (or) property type (or) operation type (or) combination of these.
 "marker" is passed by the client to resume events in case of failures.
*/
func (s *Service) StreamMlsListingEvent(in *pb.StreamMlsListingEventRequest, stream pb.MlsListingService_StreamMlsListingEventServer) error {

	ctx, span := trace.StartSpan(stream.Context(), "/listingChanges")
	defer span.End()

	md, _ := metadata.FromIncomingContext(stream.Context()) // get context from stream
	if md != nil {                                          // TODO: Metrics
		log.Printf("request has been made by %s to listen mls changes for : [%s] (empty for all changes)", md.Get("apikey"), in)
	}

	ctx, _ = context.WithDeadline(stream.Context(), time.Now().Add(time.Duration(s.Stream.DeadlineSecs)*time.Second)) //default deadline from config. if "Grpc-Timeout" is set in the header that should override.

	// get mongodb collection
	collection := s.MongoDatabase.Collection(s.ListingsCollection)

	var changeStreamPipeline primitive.A
	if in.SourceSystemKey != "" { // Subscribe to events related to a source system key.
		changeStreamPipeline = append(changeStreamPipeline, bson.D{{"fullDocument.source_system_key", in.SourceSystemKey}})
	}
	if in.PropertyType != "" { // Subscribe to events related to a property type.
		changeStreamPipeline = append(changeStreamPipeline, bson.D{{"fullDocument.property.property_type", in.PropertyType}})
	}

	var operationType primitive.A
	if in.ChangeType != "" {
		if in.ChangeType == "delete" { // subscribe to delete operation type.
			operationType = append(operationType, bson.D{{"operationType", "delete"}})
		} else { // subscribe to a specific operation type.
			operationType = append(operationType, bson.D{{"operationType", in.ChangeType}})
		}
	} else { // By default, subscribe to all operation types except delete.
		operationType = append(operationType, bson.D{{"operationType", "update"}},
			bson.D{{"operationType", "insert"}},
			bson.D{{"operationType", "replace"}})
	}
	changeStreamPipeline = append(changeStreamPipeline, bson.D{{"$or", operationType}})

	pipeline := mongo.Pipeline{bson.D{{"$match", bson.D{{"$and",
		changeStreamPipeline,
	}},
	}}}

	var changeStreamOptions options.ChangeStreamOptions
	changeStreamOptions = *options.ChangeStream().SetFullDocument(options.UpdateLookup)
	if in.Marker != "" {
		log.Printf("Resume mls events after %s", in.Marker)
		changeStreamOptions.SetResumeAfter(bsonx.Doc{{"_data", bsonx.String(in.Marker)}})
	}
	if in.ChangeStartTime != nil {
		log.Printf("Received request to get mls changes for start time: %v", in.ChangeStartTime)
		changeStartTime, err := ptypes.Timestamp(in.ChangeStartTime)
		if err != nil {
			log.Errorf("Invalid timestamp for listening to change streams ! %v", err)
			return err //TODO: Fix me
		}
		changeStreamOptions.SetStartAtOperationTime(&primitive.Timestamp{
			T: uint32(changeStartTime.UTC().Unix()),
			I: uint32(changeStartTime.UTC().UnixNano()),
		})
	}
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered panic due to stream error. %v", r)
		}
	}()

	cs, err := collection.Watch(ctx, pipeline, &changeStreamOptions)
	if cs != nil {
		defer cs.Close(ctx)
	}

	if err != nil {
		log.Errorf("Error while listening change streams ! %v", err)
		stream.Context().Err()
		// handled only for mongodb change stream error, "resume point may no longer be in the oplog". modify as specific cases other than this happens in future.
		return status.Error(codes.OutOfRange, "Unable to listen for listing changes. It may be possible that the requested change start time is out of range.")
	}

	for cs.Next(ctx) {
		result := pb.StreamMlsListingEventResponse{MlsChange: &pb.MlsChange{}}
		var eventResponse MlsEventResponse

		err := cs.Decode(&eventResponse)
		if err != nil {
			log.Errorf("unable to decode mls listings for event (id: %s). error : %v", *eventResponse.EventData.EventId, err)
			// TODO: Log error. Add Metrics. See if a partial data can be sent otherwise continue to next events.
			continue
		}

		if &eventResponse != nil {
			result.MlsListing = eventResponse.MlsListing
			result.MlsId = *eventResponse.EventDocumentKey.Id
			result.MlsChange.ChangeType = *eventResponse.EventType
			result.MlsChange.ChangeTime, _ = ptypes.TimestampProto(time.Unix(int64(eventResponse.EventTime.T), int64(eventResponse.EventTime.I))) //ignores error while parsing time.
			result.MlsChange.Marker = *eventResponse.EventData.EventId

			if err := stream.Send(&result); err != nil {
				log.Errorf("Error while streaming mls listings: %v", err)
				return err
			}
			log.Debugf("Sending mls [%s] event", result.MlsChange.ChangeType)
		}
	}
	log.Debugf("completed streaming listing changes.")
	stream.Context().Done()
	return nil
}

// aggregate mongodb pipeline
func aggregatePipelineFilter(filter *pb.MlsFilter) primitive.D {
	var pipeline primitive.D
	if filter != nil {
		if filter.PropertyType != nil && len(filter.PropertyType) > 0 {
			pipeline = append(pipeline, bson.E{Key: "property.property_type", Value: bson.M{"$in": filter.PropertyType}})
		}
		if filter.StandardStatus != nil && len(filter.StandardStatus) > 0 {
			pipeline = append(pipeline, bson.E{Key: "property.listing.standard_status", Value: bson.M{"$in": filter.StandardStatus}})
		}
		if filter.ArchitectureStyle != nil && len(filter.ArchitectureStyle) > 0 {
			pipeline = append(pipeline, bson.E{Key: "property.structure.architecture_style", Value: bson.M{"$in": filter.ArchitectureStyle}})
		}
		if filter.RdmSourceSystemKey != "" {
			pipeline = append(pipeline, bson.E{Key: "property.listing.rdm_source_system_key", Value: filter.RdmSourceSystemKey})
		}
		if filter.ListPriceMin != 0 {
			pipeline = append(pipeline, bson.E{Key: "property.listing.price.list_price", Value: bson.M{"$gte": filter.ListPriceMin}})
		}
		if filter.ListPriceMax != 0 {
			pipeline = append(pipeline, bson.E{Key: "property.listing.price.list_price", Value: bson.M{"$lte": filter.ListPriceMax}})
		}
		if filter.BedroomsMin != 0 {
			pipeline = append(pipeline, bson.E{Key: "property.structure.bedrooms_total", Value: bson.M{"$gte": filter.BedroomsMin}})
		}
		if filter.PostalCode != nil && len(filter.PostalCode) > 0 {
			pipeline = append(pipeline, bson.E{Key: "property.location.address.postal_code", Value: bson.M{"$in": filter.PostalCode}})
		}
	}
	return pipeline
}

// mongodb find options
func (s *Service) findOptions(limit int32, offset int32) *options.FindOptions {
	// mongodb find options
	findOptions := options.Find()
	findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

	actualLimit := s.getLimit(limit)
	findOptions.SetLimit(actualLimit)
	findOptions.SetSkip(int64(offset))
	return findOptions
}

func (s *Service) getLimit(limit int32) int64 {
	if limit <= 0 {
		return int64(s.Pagination.LimitDefault)
	} else if limit <= s.Pagination.LimitMax {
		return int64(limit)
	} else {
		return int64(s.Pagination.LimitMax)
	}
}

func (s *Service) SearchMlsListings(ctx context.Context, in *pb.SearchMlsListingsRequest) (*pb.SearchMlsListingsResponse, error) {

	md, _ := metadata.FromIncomingContext(ctx) // get context from stream
	if md != nil {
		log.Printf("client %s requested to search listings for input : [%s]", md.Get("apikey"), in)
	}

	err := validation.Errors{
		"standardStatus": validation.Validate(mlsvalidation.IsValidStatus(in.StandardStatus)),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.SearchMlsListingsResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var mongodbCur *mongo.Cursor
	var mongodbErr error
	if in.Q != nil {

		if in.Q.ListingId != "" {
			operator, operand := parseSearchQuery(in.Q.ListingId)
			if len(operand) < 3 {
				msg := "minimum 3 chars required to search listings"
				log.Errorf("Validation Error. %v", msg)
				return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", msg))
			}
			pipeline := s.searchPipeline("wildcard", "listing_id", operand, operator, in.Limit, in.Offset)

			opts := options.Aggregate()
			mongodbCur, mongodbErr = mongoCollection.Aggregate(ctx, pipeline, opts)
			if mongodbErr != nil {
				log.Errorf("Error while searching listings : %v", mongodbErr)
				return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching listings : %s", in))
			}
		}

	} else {
		if !in.IsRealogyListing && !in.IsLuxuryListing {
			msg := "input must be accompanied by valid search filters"
			log.Errorln(msg)
			return nil, status.Errorf(codes.InvalidArgument, msg)
		}

		pipeline := primitive.D{}
		isRealogyListing := "realogy.is_realogy_listing"
		isLuxuryListing := "realogy.is_luxury_listing"
		if in.IsRealogyListing {
			pipeline = append(pipeline, bson.E{Key: isRealogyListing, Value: aws.Bool(true)})

			if in.IsLuxuryListing {
				pipeline = append(pipeline, bson.E{Key: isLuxuryListing, Value: aws.Bool(true)})
			} else {
				var isLuxuryListingPipeline primitive.A
				isLuxuryListingPipeline = append(isLuxuryListingPipeline,
					bson.D{{isLuxuryListing, true}},
					bson.D{{isLuxuryListing, false}},
					bson.D{{isLuxuryListing, ""}},
					bson.D{{isLuxuryListing, nil}})
				pipeline = append(pipeline, bson.E{Key: "$or", Value: isLuxuryListingPipeline})
			}
		}
		if in.IsLuxuryListing {
			if !in.IsRealogyListing { // luxury listings can only be allowed to search for realogy listings
				pipeline = append(pipeline, bson.E{Key: isRealogyListing, Value: aws.Bool(true)})
			}
			pipeline = append(pipeline, bson.E{Key: isLuxuryListing, Value: aws.Bool(true)})
		}

		if in.StandardStatus != "" {
			pipeline = append(pipeline, bson.E{Key: "property.listing.standard_status", Value: in.StandardStatus})
		}

		// lastChangeTimestamp
		allowedLastChangeTs := time.Now().AddDate(0, 0, -s.BySource.AllowedLastChangeDays).UTC()
		if in.LastChangeTimestamp != nil {
			ts := timestamppb.Timestamp{Seconds: in.LastChangeTimestamp.Seconds, Nanos: 0}
			lastChangeTs := ts.AsTime()
			if lastChangeTs.Unix() > allowedLastChangeTs.Unix() {
				log.Infof("searching listings for last change timestamp : %v", lastChangeTs)
				pipeline = append(pipeline, bson.E{Key: "last_change_date", Value: bson.M{"$gte": lastChangeTs}})
			} else {
				log.Errorf("listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays)
				return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays))
			}
		}
		//else { // default
		//	pipeline = append(pipeline, bson.E{Key: "last_change_date", Value: bson.M{"$gte": allowedLastChangeTs}})
		//}

		// mongodb find options
		findOptions := s.findOptions(in.Limit, in.Offset)
		findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

		mongodbCur, mongodbErr = mongoCollection.Find(ctx, &pipeline, findOptions)
		if mongodbErr != nil {
			log.Errorf("Error while processing the request to search mls: %v", mongodbErr)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching listings for input : %s", in))
		}
	}

	if mongodbCur == nil {
		msg := "unable to find listings due to internal error"
		log.Errorf("%s. mongodb cursor is nil", msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	// iterate mongo cursor and create response
	for mongodbCur.Next(ctx) {
		var result pb.MlsListing
		err := mongodbCur.Decode(&result)
		if err != nil {
			log.Errorf("unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		msg := fmt.Sprintf("unable to find listings for input : %s", in)
		log.Errorln(msg)
		return nil, status.Errorf(codes.NotFound, msg)
	}
	return response, nil
}

func (s *Service) GetRealogyListings(ctx context.Context, in *pb.RealogyListingsRequest) (*pb.RealogyListingsResponse, error) {

	md, _ := metadata.FromIncomingContext(ctx) // get context from stream
	if md != nil {
		log.Printf("client %s requested to Realogy listings for input : [%s]", md.Get("apikey"), in)
	}

	err := validation.Errors{
		"standardStatus": validation.Validate(in.StandardStatus, validation.In("ACTIVE", "INACTIVE")),
	}.Filter()

	if err != nil {
		log.Errorf("Validation Error. %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", err))
	}

	response := &pb.RealogyListingsResponse{}

	// get mongodb collection
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	var mongodbCur *mongo.Cursor
	var mongodbErr error
	if in.Q != nil {

		if in.Q.ListingId != "" {
			operator, operand := parseSearchQuery(in.Q.ListingId)
			if len(operand) < 3 {
				msg := "minimum 3 chars required to search listings"
				log.Errorf("Validation Error. %v", msg)
				return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid input. %v", msg))
			}
			pipeline := s.searchPipeline("wildcard", "listing_id", operand, operator, in.Limit, in.Offset)

			opts := options.Aggregate()
			mongodbCur, mongodbErr = mongoCollection.Aggregate(ctx, pipeline, opts)
			if mongodbErr != nil {
				log.Errorf("Error while searching listings : %v", mongodbErr)
				return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching listings : %s", in))
			}
		}

	} else {

		pipeline := primitive.D{}
		isRealogyListing := "realogy.is_realogy_listing"

		pipeline = append(pipeline, bson.E{Key: isRealogyListing, Value: aws.Bool(true)})

		if in.StandardStatus != "" {
			pipeline = append(pipeline, bson.E{Key: "property.listing.standard_status", Value: in.StandardStatus})
		}
		if in.SourceSystemKey != "" {
			pipeline = append(pipeline, bson.E{Key: "property.listing.source_system_key", Value: in.SourceSystemKey})
		}

		// lastChangeTimestamp
		allowedLastChangeTs := time.Now().AddDate(0, 0, -s.BySource.AllowedLastChangeDays).UTC()
		if in.LastChangeTimestamp != nil {
			ts := timestamppb.Timestamp{Seconds: in.LastChangeTimestamp.Seconds, Nanos: 0}
			lastChangeTs := ts.AsTime()
			if lastChangeTs.Unix() > allowedLastChangeTs.Unix() {
				log.Infof("searching listings for last change timestamp : %v", lastChangeTs)
				pipeline = append(pipeline, bson.E{Key: "last_change_date", Value: bson.M{"$gte": lastChangeTs}})
			} else {
				log.Errorf("listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays)
				return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Listings cannot be searched beyond last %v days", s.BySource.AllowedLastChangeDays))
			}
		}
		// mongodb find options
		findOptions := s.findOptions(in.Limit, in.Offset)
		findOptions.SetMaxTime(time.Duration(s.MaxQueryTimeSecs) * time.Second)

		mongodbCur, mongodbErr = mongoCollection.Find(ctx, &pipeline, findOptions)
		if mongodbErr != nil {
			log.Errorf("Error while processing the request to search mls: %v", mongodbErr)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error while searching listings for input : %s", in))
		}
	}

	if mongodbCur == nil {
		msg := "unable to find listings due to internal error"
		log.Errorf("%s. mongodb cursor is nil", msg)
		return nil, status.Errorf(codes.Internal, msg)
	}
	// iterate mongo cursor and create response
	for mongodbCur.Next(ctx) {
		var result pb.MlsListing
		err := mongodbCur.Decode(&result)
		if err != nil {
			log.Errorf("unable to decode the document: %v", err)
		}
		response.MlsListings = append(response.MlsListings, &result)
	}

	if len(response.MlsListings) == 0 {
		msg := fmt.Sprintf("unable to find listings for input : %s", in)
		log.Errorln(msg)
		return nil, status.Errorf(codes.NotFound, msg)
	}
	return response, nil
}

func parseSearchQuery(query string) (string, string) {
	if strings.Contains(query, ":") {
		log.Println("search query :", query)
		searchParams := strings.Split(query, ":")

		if len(searchParams) == 2 {
			return searchParams[0], searchParams[1]
		} else if len(searchParams) == 1 {
			return searchParams[0], ""
		}
	}
	return "", ""
}

func (s *Service) searchPipeline(mongodbOperator string, path string, query string, operator string, limit int32, offset int32) primitive.A {
	var searchQuery string
	//pipeline := primitive.A{}
	log.Printf("search pipeline - mongodbOperator: [%v], path: [%v], query: [%v], operator: [%v]", mongodbOperator, path, query, operator)
	if mongodbOperator == "wildcard" {
		if operator == pb.ComparisonOperators_like.String() {
			searchQuery = "*" + query + "*"
		} else {
			searchQuery = query
		}
		log.Printf("search query: %s", searchQuery)

		return bson.A{
			bson.D{{"$search",
				bson.D{
					bson.E{Key: "index", Value: "listingIdSearchIdx"},
					bson.E{Key: "wildcard",
						Value: bson.D{
							{"path", path},
							{"query", searchQuery},
							{"allowAnalyzedField", true},
						},
					}},
			}},
			bson.D{{Key: "$limit", Value: s.getLimit(limit)}},
			bson.D{{Key: "$skip", Value: offset}},
		}
	}
	return nil
}

// health check api
func (s *Service) HealthCheck(ctx context.Context, in *pb.HealthRequest) (*pb.HealthResponse, error) {
	response := &pb.HealthResponse{}
	mongoCollection := s.MongoDatabase.Collection(s.ListingsCollection)

	runCmdOpts := &options.RunCmdOptions{ReadPreference: readpref.Nearest(readpref.WithMaxStaleness(90 * time.Second))}

	err := mongoCollection.Database().RunCommand(ctx, bsonx.Doc{{"ping", bsonx.String("1")}}, runCmdOpts).Decode(&response)
	if err != nil {
		log.Errorf("Error while pinging mongodb for health check : %v", err)
		return nil, err
	}
	return response, nil
}
