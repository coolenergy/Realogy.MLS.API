package transformer

import (
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"mlslisting/internal/mlsvalidation"
	"mlslisting/internal/models"
	"time"
)

func TransformUpdateListingInputToUpdateMlsListing(in *pb.UpdateMlsListingByListingIdRequest) models.UpdateMLSListing {
	var standardStatus string
	var listPrice float64
	var closePrice float64
	currentTime := time.Now()
	var listingContractDate time.Time
	var expirationDate time.Time
	var cancelDate time.Time
	var pendingDate time.Time
	var closeDate time.Time
	if !mlsvalidation.IsEmpty(in.Property) && !mlsvalidation.IsEmpty(in.Property.Listing) {
		standardStatus = in.Property.Listing.StandardStatus
		if !mlsvalidation.IsEmpty(in.Property.Listing.Price) {
			listPrice = in.Property.Listing.Price.ListPrice
			closePrice = in.Property.Listing.Price.ClosePrice
		}
		if !mlsvalidation.IsEmpty(in.Property.Listing.Dates) {
			if !mlsvalidation.IsEmpty(in.Property.Listing.Dates.ListingContractDate) {
				listingContractDate = in.Property.Listing.Dates.ListingContractDate.AsTime()
			}
			if !mlsvalidation.IsEmpty(in.Property.Listing.Dates.ExpirationDate) {
				expirationDate = in.Property.Listing.Dates.ExpirationDate.AsTime()
			}
			if !mlsvalidation.IsEmpty(in.Property.Listing.Dates.CancellationDate) {
				cancelDate = in.Property.Listing.Dates.CancellationDate.AsTime()
			}
			if !mlsvalidation.IsEmpty(in.Property.Listing.Dates.PendingTimestamp) {
				pendingDate = in.Property.Listing.Dates.PendingTimestamp.AsTime()
			}
			if !mlsvalidation.IsEmpty(in.Property.Listing.Dates.CloseDate) {
				closeDate = in.Property.Listing.Dates.CloseDate.AsTime()
			}
		}
	}

	return models.UpdateMLSListing{
		LastChangeDate: currentTime,
		Property: &models.UpdateProperty{
			Listing: models.UpdateListing{
				StandardStatus: standardStatus,
				Price: &models.UpdatePrice{
					ListPrice:  listPrice,
					ClosePrice: closePrice,
				},
				Dates: &models.UpdateDates{
					LastChangeDate:      currentTime,
					ListingContractDate: listingContractDate,
					ExpirationDate:      expirationDate,
					CancellationDate:    cancelDate,
					PendingTimestamp:    pendingDate,
					CloseDate:           closeDate,
				},
			},
		},
	}
}

func TransformListingInputToMlsListing(in *pb.MlsListingInput) models.MLSListing {
	return models.MLSListing{
		Property: &models.Property{
			PropertyType: in.Property.PropertyType,
			Listing: models.Listing{
				ListingId:          in.ListingId,
				SourceSystemKey:    in.RdmSourceSystemKey,
				RdmSourceSystemKey: in.RdmSourceSystemKey,
				StandardStatus:     in.Property.Listing.StandardStatus,
				Price: &models.Price{
					ListPrice:  in.Property.Listing.Price.ListPrice,
					ClosePrice: in.Property.Listing.Price.ClosePrice,
					Currency:   in.Property.Listing.Price.Currency,
				},
				Dates: &models.Dates{
					LastChangeDate: time.Now(),
					CloseDate:      in.Property.Listing.Dates.CloseDate.AsTime(),
					InsertedDate:   time.Now(),
				},
			},
			Location: models.Location{
				Address: &models.Address{
					UnparsedAddress:     in.Property.Location.Address.UnparsedAddress,
					City:                in.Property.Location.Address.City,
					CountyOrParish:      in.Property.Location.Address.CountyOrParish,
					StateOrProvince:     in.Property.Location.Address.StateOrProvince,
					Country:             in.Property.Location.Address.Country,
					InternationalRegion: in.Property.Location.Address.InternationalRegion,
				},
			},
		},
	}
}
