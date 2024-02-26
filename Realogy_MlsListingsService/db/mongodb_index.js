// to be deleted
db.listings.createIndex({"listing_id" : 1}, {"name" : "listingIdIndex"})

db.listings.createIndex({ "listing_id": 1}, { "name" : "listingIdIdxWithCollation", collation: { locale: "en", strength: 2 } })

db.listings.createIndex({"source_system_key" : 1}, {"name" : "sourceSystemKeyIndex"})

db.listings.createIndex({"dash.listing_guid" : 1}, {"name" : "listingGuidIndex"})

db.listings.createIndex({"property.listing.agent_office.list_agent.list_agent_mls_id" : 1}, {"name" : "listAgentMlsIdIndex"})

db.listings.createIndex({"dash.listing_agent_guid" : 1}, {"name" : "listingAgentGuidIndex"})

db.listings.createIndex({"master_id.list_agent_master_id" : 1}, {"name" : "listAgentMasteridIndex"})

db.listings.createIndex({"master_id.list_office_master_id":1}, { collation: { locale: "en", strength: 2 } } )

db.listings.createIndex({"master_id.company_master_id":1}, { collation: { locale: "en", strength: 2 } } )

db.listings.createIndex({ "property.location.address.city": 1, "property.location.address.state_or_province": 1 }, { collation: { locale: "en", strength: 2 } })

db.listings.createIndex({"property.location.address.postal_code" : 1}, {"name" : "postalCodeIndex"})

db.listings.createIndex({ _source: 1, "property.listing.standard_status": 1, "property.property_type": 1}, {name: "index for retrieval by status, source and property type"})

db.listings.createIndex({ "dash.company_staff_guid": 1}, { collation: { locale: "en", strength: 2 } })

db.listings.createIndex({"source_system_key" : 1, "last_change_date" : 1}, {"name" : "sourceSystemKeyLastChangeDateIndex"})

// indexes for realogy based fields.
db.listings.createIndex({"realogy.is_realogy_listing" : 1, "realogy.is_luxury_listing" : 1, "property.listing.standard_status": 1, "last_change_date" : 1}, {"name" : "realogyListingsPartialIndex"}, {"partialFilterExpression" : {"realogy.is_realogy_listing" : true, "realogy.is_luxury_listing" : true}})

// search indexes
{
    "analyzer": "lucene.standard",
    "searchAnalyzer": "lucene.standard",
    "mappings": {
    "dynamic": false,
        "fields": {
        "listing_id": {
            "type": "string",
                "analyzer": "lucene.keyword"
        }
    }
}
}