package services

import (
	"errors"
	"github.com/jinzhu/now"
	_ "github.com/jinzhu/now"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"time"
)

// mongodb pipeline that queries sold listings (standardStatus="SOLD")
func SoldListingsPipeline(in *pb.GetMlsSoldListingsRequest, pipeline primitive.D) (primitive.D, error) {

	if in.StartDate != "" && in.EndDate != "" { // both start and end date range is given
		startDate, err1 := now.Parse(in.StartDate)
		endDate, err2 := now.Parse(in.EndDate)
		if err1 != nil || err2 != nil {
			return nil, errors.New("date range should be in YYYY-MM-DD format")
		}
		dateRange := startDate.AddDate(0, 6, 0)
		if dateRange.Sub(endDate) < 0 {
			return nil, errors.New("date range for sold listings should be within 6 months")
		}
		pipeline = append(pipeline, bson.E{Key: "property.listing.dates.close_date", Value: bson.M{"$gte": startDate, "$lte": endDate}})

	} else if in.StartDate != "" { // only start date is given in the input. the end date defaults to 6 months from start date.
		startDate, err := now.Parse(in.StartDate)
		if err != nil {
			return nil, errors.New("date range should be in YYYY-MM-DD format")
		}
		pipeline = append(pipeline, bson.E{Key: "property.listing.dates.close_date", Value: bson.M{"$gte": startDate, "$lte": time.Now().AddDate(0, 6, 0)}})
	} else if in.StartDate == "" && in.EndDate == "" { // both start and end date range is not given. defaults to last 6months.
		y,m,d := time.Now().Date()
		pipeline = append(pipeline, bson.E{Key: "property.listing.dates.close_date", Value: bson.M{"$gte": time.Date(y,m-6,d,0,0,0,0,time.UTC), "$lte": time.Date(y,m,d,0,0,0,0,time.UTC)}})
	} else {
		return nil, errors.New("start date is required")
	}

	return append(pipeline, bson.E{Key: "property.listing.standard_status", Value: "SOLD"}), nil
}
