package mlsvalidation

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"reflect"
)

type StandardStatus int64

const (
	CurrencyUSD = "USD"
)

const (
	_ StandardStatus = iota
	Active
	Inactive
	Sold
	Canceled
	Hold
	Unknown
	Expired
	Temp
	Terminated
	Pending
	Withdrawn
)

func (s StandardStatus) String() string {
	switch s {
	case Active:
		return "ACTIVE"
	case Inactive:
		return "INACTIVE"
	case Sold:
		return "SOLD"
	case Canceled:
		return "CANCELED"
	case Hold:
		return "HOLD"
	case Unknown:
		return "UNKNOWN"
	case Expired:
		return "EXPIRED"
	case Temp:
		return "TEMP"
	case Terminated:
		return "TERMINATED"
	case Pending:
		return "PENDING"
	case Withdrawn:
		return "WITHDRAWN"
	default:
		return ""
	}
}

func IsValidStatus(status string) bool {
	for i := Active; i <= Withdrawn; i++ {
		if i.String() == status {
			return true
		}
	}
	return false
}

func IsEmpty(input interface{}) bool {
	switch input.(type) {
	case string:
		return input == ""
	case int, int64, int32, int16, int8, float32, float64:
		return input == 0
	case timestamppb.Timestamp:
		return input == nil || input.(*timestamppb.Timestamp).AsTime().IsZero()
	default:
		return input == nil || reflect.ValueOf(input).IsNil()
	}
}

// ValidateInsertListing : validate a new listing for insert.
func ValidateInsertListing(in *pb.MlsListingInput) error {
	//  Address not null or empty except LAND property type.
	if in.Property.PropertyType != "LAND" && (IsEmpty(in.Property.Location.Address) || IsEmpty(in.Property.Location.Address.UnparsedAddress)) {
		return status.Errorf(codes.InvalidArgument, "Unparsed Address can not be empty or null")
	}
	if IsEmpty(in.Property.Location.Address.City) {
		return status.Errorf(codes.InvalidArgument, "City can not be empty or null")
	}
	if (in.Property.Location.Address.Country == "USA") && (IsEmpty(in.Property.Location.Address.CountyOrParish) || IsEmpty(in.Property.Location.Address.StateOrProvince)) {
		return status.Errorf(codes.InvalidArgument, "County/Parish and State/Province can not be nil for USA.")
	}
	if IsEmpty(in.Property.Location.Address.Country) {
		return status.Errorf(codes.InvalidArgument, "Country can not be empty or null")
	}
	if in.Property.Listing.Price.ListPrice < 0 {
		return status.Errorf(codes.InvalidArgument, "List Price invalid.")
	}
	if IsEmpty(in.Property.Listing.Price.Currency) {
		in.Property.Listing.Price.Currency = CurrencyUSD
	}
	if !IsValidStatus(in.Property.Listing.StandardStatus) {
		return status.Errorf(codes.InvalidArgument, "Standard Status value was not recognized as an acceptable value : "+in.Property.Listing.StandardStatus)
	}
	if in.Property.Listing.StandardStatus == "SOLD" && (IsEmpty(in.Property.Listing.Dates) || IsEmpty(in.Property.Listing.Dates.CloseDate) || IsEmpty(in.Property.Listing.Price) || IsEmpty(in.Property.Listing.Price.ClosePrice)) {
		return status.Errorf(codes.InvalidArgument, "Closed Date and Close Price required for SOLD listing")
	}
	return nil
}

// ValidateUpdateListing : validate an update to a listing.
func ValidateUpdateListing(in *pb.UpdateMlsListingByListingIdRequest, listingFromDB *pb.MlsListing) error {
	if !((in.SourceSystemKey == "SOLO" || in.SourceSystemKey == "ELL") || (!IsEmpty(listingFromDB.Realogy) &&
		listingFromDB.Realogy.IsRealogyListing && !IsEmpty(listingFromDB.MasterId) && !IsEmpty(listingFromDB.MasterId.ListAgentMasterId) &&
		!IsEmpty(listingFromDB.MasterId.ListOfficeMasterId) &&
		!IsEmpty(listingFromDB.MasterId.CompanyMasterId))) {
		return status.Errorf(codes.Unauthenticated, "The request to update this listing is not authorized due to invalid source,missing master ids, or is not flagged as a realogy listing")
	}
	if in.Property.Listing != nil {
		// if the status is empty we can assume they don't want to update it and can keep the current value
		if !IsEmpty(in.Property.Listing.StandardStatus) && !IsValidStatus(in.Property.Listing.StandardStatus) {
			return status.Errorf(codes.InvalidArgument, "Standard Status value was not recognized as an acceptable value : "+in.Property.Listing.StandardStatus)
		}
		// if the price is empty we can assume they don't want to update it and can keep the current value
		if !IsEmpty(in.Property.Listing.Price) && !IsEmpty(in.Property.Listing.Price.ListPrice) && in.Property.Listing.Price.ListPrice < 0 {
			return status.Errorf(codes.InvalidArgument, "List Price invalid.")
		}
		if in.Property.Listing.StandardStatus == "SOLD" &&
			(IsEmpty(in.Property.Listing.Dates) || IsEmpty(in.Property.Listing.Dates.CloseDate) || IsEmpty(in.Property.Listing.Price) || IsEmpty(in.Property.Listing.Price.ClosePrice) || in.Property.Listing.Dates.CloseDate.AsTime().Before(listingFromDB.Property.Listing.Dates.ListingContractDate.AsTime())) {
			return status.Errorf(codes.InvalidArgument, "Closed Date and Close Price required for SOLD listing and Close date must be later than ListingContractDate")
		}
		if in.Property.Listing.StandardStatus == "CANCELED" &&
			(IsEmpty(in.Property.Listing.Dates) || IsEmpty(in.Property.Listing.Dates.CancellationDate) || in.Property.Listing.Dates.CancellationDate.AsTime().IsZero() || in.Property.Listing.Dates.CancellationDate.AsTime().Before(listingFromDB.Property.Listing.Dates.ListingContractDate.AsTime())) {
			return status.Errorf(codes.InvalidArgument, "Cancellation Date required for CANCELED listing and Cancellation date must be later than ListingContractDate")
		}
		if in.Property.Listing.StandardStatus == "PENDING" &&
			(IsEmpty(in.Property.Listing.Dates) || IsEmpty(in.Property.Listing.Dates.PendingTimestamp) || in.Property.Listing.Dates.PendingTimestamp.AsTime().Before(listingFromDB.Property.Listing.Dates.ListingContractDate.AsTime())) {
			return status.Errorf(codes.InvalidArgument, "Pending Timestamp required for PENDING listing and Pending Timestamp must be later than ListingContractDate")
		}
		if in.Property.Listing.StandardStatus == "EXPIRED" &&
			(IsEmpty(in.Property.Listing.Dates) || IsEmpty(in.Property.Listing.Dates.ExpirationDate) || in.Property.Listing.Dates.ExpirationDate.AsTime().IsZero() || in.Property.Listing.Dates.ExpirationDate.AsTime().Before(listingFromDB.Property.Listing.Dates.ListingContractDate.AsTime())) {
			return status.Errorf(codes.InvalidArgument, "Expiration Date required for EXPIRED listing and expiration date must be later than ListingContractDate")
		}
	}
	return nil
}
