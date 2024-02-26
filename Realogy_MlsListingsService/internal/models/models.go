package models

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

type MLSListing struct {
	ListingID          string `bson:"listing_id"`
	RdmSourceSystemKey string `bson:"rdm_source_system_key"`
	SourceSystemKey    string `bson:"source_system_key"`
	// The fields and groups contained within the Property Group.
	Property *Property `bson:"property,omitempty"`
	// The Media type is a representation of media, such as photos, virtual tours, documents/supplements, etc.
	Media *Media `bson:"media,omitempty"`
	// The OpenHouse type is a collection of fields commonly used to record an open house event.
	OpenHouse      *OpenHouse `bson:"open_house,omitempty"`
	LastChangeDate time.Time  `bson:"last_change_date"`
	InsertedBy     string     `bson:"inserted_by"`
}

type UpdateMLSListing struct {
	// The fields and groups contained within the Property Group.
	Property *UpdateProperty `bson:"property,omitempty"`
	// The Media type is a representation of media, such as photos, virtual tours, documents/supplements, etc.
	Media *UpdateMedia `bson:"media,omitempty"`
	// The OpenHouse type is a collection of fields commonly used to record an open house event.
	OpenHouse      *UpdateOpenHouse `bson:"open_house,omitempty"`
	LastChangeDate time.Time        `bson:"last_change_date"`
}

type UpdateProperty struct {
	Listing UpdateListing `bson:"listing,omitempty"`
}

type Property struct {
	PropertyType string   `bson:"property_type"`
	Listing      Listing  `bson:"listing, omitempty"`
	Location     Location `bson:"location"`
}

type UpdateListing struct {
	StandardStatus string         `bson:"standard_status,omitempty"`
	Price          *UpdatePrice   `bson:"price,omitempty"`
	Dates          *UpdateDates   `bson:"dates,omitempty"`
	Remarks        *UpdateRemarks `bson:"remarks,omitempty"`
}

type Listing struct {
	ListingId          string   `bson:"listing_id"`
	SourceSystemKey    string   `bson:"source_system_key"`
	StandardStatus     string   `bson:"standard_status"`
	Price              *Price   `bson:"price"`
	Dates              *Dates   `bson:"dates"`
	RdmSourceSystemKey string   `bson:"rdm_source_system_key"`
	Remarks            *Remarks `bson:"remarks,omitempty"`
}

type Location struct {
	Address *Address `bson:"address"`
}

type Address struct {
	UnparsedAddress     string `bson:"unparsed_address"`
	City                string `bson:"city"`
	CountyOrParish      string `bson:"county_or_parish"`
	StateOrProvince     string `bson:"state_or_province"`
	Country             string `bson:"country"`
	InternationalRegion string `bson:"international_region"`
}

type UpdatePrice struct {
	ListPrice  float64 `bson:"list_price,omitempty"`
	ClosePrice float64 `bson:"close_price,omitempty"`
}

type Price struct {
	ListPrice  float64 `bson:"list_price"`
	ClosePrice float64 `bson:"close_price"`
	Currency   string  `bson:"currency"`
}

type UpdateDates struct {
	LastChangeDate      time.Time `bson:"last_change_date,omitempty"`
	ListingContractDate time.Time `bson:"listing_contract_date,omitempty"`
	ExpirationDate      time.Time `bson:"expiration_date,omitempty"`
	CloseDate           time.Time `bson:"close_date,omitempty"`
	CancellationDate    time.Time `bson:"cancellation_date,omitempty"`
	PendingTimestamp    time.Time `bson:"pending_timestamp,omitempty"`
}

type Dates struct {
	LastChangeDate time.Time `bson:"last_change_date"`
	InsertedDate   time.Time `bson:"inserted_date"`
	CloseDate      time.Time `bson:"close_date"`
}

type UpdateRemarks struct {
	PublicRemarks   string `bson:"public_remarks,omitempty"`
	PrivateRemarks  string `bson:"private_remarks,omitempty"`
	SellingComments string `bson:"selling_comments,omitempty"`
}

type Remarks struct {
	PublicRemarks   string `bson:"public_remarks"`
	PrivateRemarks  string `bson:"private_remarks"`
	SellingComments string `bson:"selling_comments"`
}

type UpdateMedia struct {
	NumImages             int32                `bson:"num_images,omitempty"`
	ModificationTimestamp *timestamp.Timestamp `bson:"modification_timestamp,omitempty"`
	// need to revisit if this needs to be recalculated when images are changed.
	ImageHashCode       string               `bson:"image_hash_code,omitempty"`
	Uuid                string               `bson:"uuid,omitempty,omitempty"`
	LastChangeTimestamp *timestamp.Timestamp `bson:"last_change_timestamp,omitempty"`
	MediaInfo           []*MediaInfo         `bson:"media_info,omitempty"`
}

type Media struct {
	NumImages             int32                `bson:"num_images"`
	ModificationTimestamp *timestamp.Timestamp `bson:"modification_timestamp"`
	ImageHashCode         string               `bson:"image_hash_code"`
	Uuid                  string               `bson:"uuid"`
	LastChangeTimestamp   *timestamp.Timestamp `bson:"last_change_timestamp"`
	MediaInfo             []*MediaInfo         `bson:"media_info"`
}

type MediaInfo struct {
	IndexNum              int32                `bson:"index_num"`
	MediaUrl              string               `bson:"media_url"`
	PhotosChangeTimestamp *timestamp.Timestamp `bson:"photos_change_timestamp"`
	ImageWidth            int32                `bson:"image_width"`
	ImageHeight           int32                `bson:"image_height"`
	Md5                   string               `bson:"md5"`
}

type UpdateOpenHouse struct {
	IsOpenHomes bool         `bson:"is_open_homes,omitempty"`
	OpenHomes   []*OpenHomes `bson:"open_homes,omitempty"`
}

type OpenHouse struct {
	IsOpenHomes bool         `bson:"is_open_homes"`
	OpenHomes   []*OpenHomes `bson:"open_homes"`
}

type OpenHomes struct {
	HashCode               string               `bson:"hash_code"`
	OpenHouseDate          *timestamp.Timestamp `bson:"open_house_date"`
	OpenHouseStartTime     *timestamp.Timestamp `bson:"open_house_start_time"`
	OpenHouseEndTime       *timestamp.Timestamp `bson:"open_house_end_time"`
	OriginalEntryTimestamp *timestamp.Timestamp `bson:"original_entry_timestamp"`
	ModificationTimestamp  *timestamp.Timestamp `bson:"modification_timestamp"`
	OpenHouseRemarks       string               `bson:"open_house_remarks"`
	IsAppointmentNeeded    bool                 `bson:"is_appointment_needed"`
}
