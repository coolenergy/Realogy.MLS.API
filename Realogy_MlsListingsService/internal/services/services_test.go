package services_test

import (
	"encoding/json"
	"errors"
	"fmt"
	pb "mlslisting/internal/generated/realogy.com/api/mls/v1"
	"mlslisting/internal/services/mock"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// The grpc response is in "snake_case" but this should be transformed to "camelCase" by grpc gateway.
	listing100018           = string(`{"mls_listings":[{"property":{"property_type":"LAND","property_sub_type":"","financial":null,"listing":{"listing_id":"100018","source_system_key":"KY_WKRMLS","source_system_name":"","standard_status":"ACTIVE","building_permits":"","documents_available":"","disclosures":"","contract":{"current_financing":"","special_listing_conditions":{"is_foreclosure":false,"is_short_sale":false,"is_probate_sale":false}},"price":{"list_price":15000,"list_price_high":0,"is_price_reduced":true,"price_change_timestamp":null,"original_list_price":19000,"close_price":0,"pet_rent":0,"months_rent_upfront":0},"agent_office":{"list_agent":{"list_agent_fullname":"JOEY RICKS","list_agent_mls_id":"21","list_agent_office_phone":"","list_agent_office_phone_type":"","list_agent_state_license":"","list_agent_state_license_state":"","list_agent_email":"","list_agent_active":false,"list_agent_address":"","list_agent_city":"","list_agent_state_or_province":"","list_agent_postal_code":"","list_agent_type":"","list_agent_original_entry_timestamp":null,"list_agent_modification_timestamp":null},"list_office":{"list_office_name":"COLDWELL BANKER SERVICE 1ST REALTY","list_office_phone":"800-522-4699","list_office_mls_id":"23","list_office_address":"","list_office_city":"","list_office_state_or_province":"","list_office_postal_code":"","list_office_email":"","list_office_fax":"","list_office_original_entry_timestamp":null,"list_office_modification_timestamp":null},"co_list_agent":null,"co_list_office":null,"buyer_agent":null,"buyer_office":null,"co_buyer_agent":null,"co_buyer_office":null},"compensation":{"list_agency_compensation":null,"buyer_agency_compensation":{"percentage":"","fee":""}},"dates":{"listing_contract_date":null,"first_appeared_date":null,"expiration_date":null,"last_change_date":null,"status_change_date":null,"inserted_date":null,"year_converted":0,"age":0,"original_entry_timestamp":null,"close_date":null,"cancellation_date":null,"pending_timestamp":null,"on_market_date":null,"contingent_date":null,"off_market_date":null,"cumulative_days_on_market":0,"modification_timestamp":null},"remarks":{"public_remarks":"Nice corner lot at Arrow Head Golf Course. This lot is ready to be built on.","private_remarks":"","misc_info":"","selling_comments":""},"marketing":{"virtual_tour_url_unbranded":"","internet_automated_valuation_display":true,"internet_consumer_comment":true,"internet_entire_listing_display":false,"internet_address_display":false,"website_detail_page_url":""},"closing":null,"home_warranty":false},"tax":{"zoning":"","parcel_number":"","tax_tract":"Cherokee Hil","tax_annual_amount":0,"tax_year":0,"tax_other_annual_assessment_amount":0},"hoa":{"association_fee":0,"association_fee_frequency":0,"association_fee2":0,"association_fee2_frequency":0,"association_amenities":"","association_fee_includes":"","pets_allowed":"","association_name":""},"location":{"gis":{"cross_street":"","map_coordinate":"","directions":"","latitude":36.8587035,"longitude":-87.8069197,"mls_latitude":36.8287,"mls_longitude":-87.87515708,"parcel_latitude":0,"parcel_longitude":0,"geocoded_city":"Cadiz","neighborhood_id":""},"address":{"unparsed_address":"000 Comanche","city":"Cadiz","county_or_parish":"Trigg","state_or_province":"KY","postal_code":"42211","street_dir_prefix":"","street_dir_suffix":"","street_name":"","street_number":"","unit_number":"","township":""},"area":{"mls_area_major":"","subdivision_name":""},"school":{"school_district":"","elementary_school":"","elementary_school_district":"","middle_or_junior_school":"Trigg County","high_school":"Trigg County","high_school_district":"","middle_or_junior_school_district":"","sd_unif_id":"2105580","sd_elem_id":"","sd_sec_id":"","saz_elem_id":"KY-TRIGG CO-5702-KY-PB25665","saz_middle_id":"KY-TRIGG CO-5702-KY-PB25663","saz_high_id":"KY-TRIGG CO-5702-KY-PB25664"}},"structure":{"architecture_style":"","heating":"","cooling":"","construction_materials":"","flooring":"","levels":"","bedrooms_total":0,"bathrooms_full":0,"bathrooms_partial":0,"bathrooms_onequarter":0,"bathrooms_threequarter":0,"bathrooms_half":0,"building_name":"","building_features":"","building_area_total":0,"building_area_source":"","living_area":0,"roof":"","parking_features":"","parking_total":"","other_parking":"","other_parking_spaces":"","garage_spaces":0,"carport_spaces":0,"covered_spaces":0,"open_parking_spaces":0,"year_built":0,"stories_total":0,"fireplace":false,"fireplace_features":"","fireplace_total":0,"door_features":"","foundation_details":"","insulation_desc":"","room_type":"","builder_name":"","accessibility_features":"","flood_area":"","homeowners_prot_plan":"","exterior_features":"","interior_features":"","body_type":"","basement":"","other_structures":"","builder_model":"","structure_type":"","rooms":{"rooms_total":0,"kitchen_dim":"","living_rm_dim":"","master_br_dim":"","dining_rm_dim":"","family_rm_dim":"","dining_desc":"","family_room_desc":"","kitchen_desc":"","living_rm_desc":"","bedroom_desc":"","bathroom_desc":""},"property_condition":{"is_fixer_upper":false,"is_new_construction":false},"approximate_office_squarefeet":0,"approximate_retail_squarefeet":0,"approximate_warehouse_squarefeet":0,"total_restrooms":0,"load":""},"characteristics":{"lot_size_acres":"","lot_size_dimensions":"","lot_features":"","lot_size_square_feet":8712,"pool_features":"","private_pool":false,"view":"","laundry_features":"","spa_features":"","community_features":"","complex_name":"","number_of_units_in_community":0,"water_body_name":"","water_front_features":"","water_front":false,"frontage_type":"","number_of_units_total":0,"hide_from_prelogin_search":false,"senior_community":false,"is_smart_home":false,"current_use":"","possible_use":"","number_of_lots":0,"number_of_pads":0,"development_status":"","fencing":"","road_surface_type":"","road_responsibility":"","misc_utilities_desc":"","furnished":"","leaseterm":"","is_renters_insurance_required":false},"utilities":{"water_source":"","sewer":"","utilities":"","number_of_separate_electricmeters":0,"number_of_separate_gasmeters":0,"number_of_separate_watermeters":0},"equipment":{"other_equipment":"","appliances":"","security_features":""},"business":{"ownership_type":"","lease_amount_frequency":""}},"media":null,"open_house":null,"dash":null,"master_id":null}]}`)
	listing29071815WithDash = string(`{"mls_listings":[{"property":{"property_type":"Single Family","property_sub_type":"Twin/Semi-Detached","photos_count":15,"financial":{"rent_includes":"","sales_includes":"","electric_expense":0,"tenant_pays":"","owner_pays":"","income_includes":"","is_rent_control":false,"total_actual_rent":0},"listing":{"listing_id":"29071815","source_system_key":"BRIGHTMLS","source_system_name":"","standard_status":"ACTIVE","building_permits":"","documents_available":"","disclosures":"","contract":{"current_financing":"","special_listing_conditions":{"is_foreclosure":false,"is_short_sale":false,"is_probate_sale":false}},"price":{"list_price":115000,"list_price_high":0,"is_price_reduced":false,"price_change_timestamp":null,"original_list_price":0,"close_price":0,"pet_rent":0,"months_rent_upfront":0},"agent_office":{"list_agent":{"list_agent_fullname":"Anne Fitzgerald","list_agent_mls_id":"3172336","list_agent_office_phone":"2159387800","list_agent_office_phone_type":"","list_agent_state_license":"","list_agent_state_license_state":"","list_agent_email":"","list_agent_active":false,"list_agent_address":"","list_agent_city":"","list_agent_state_or_province":"","list_agent_postal_code":"","list_agent_type":"","list_agent_original_entry_timestamp":null,"list_agent_modification_timestamp":null},"list_office":{"list_office_name":"Better Homes Realty Group","list_office_phone":"","list_office_mls_id":"55916","list_office_address":"","list_office_city":"","list_office_state_or_province":"","list_office_postal_code":"","list_office_email":"","list_office_fax":"","list_office_original_entry_timestamp":null,"list_office_modification_timestamp":null},"co_list_agent":{"co_list_agent_full_name":"","co_list_agent_mls_id":"","co_list_agent_office_phone":""},"co_list_office":{"co_list_office_name":"","co_list_office_mls_id":"","co_list_office_phone":""},"buyer_agent":{"buyer_agent_fullname":"","buyer_agent_mls_id":"","buyer_office_phone":""},"buyer_office":null,"co_buyer_agent":null,"co_buyer_office":null},"compensation":null,"dates":{"listing_contract_date":"2019-03-07T00:00:00Z","first_appeared_date":null,"expiration_date":null,"last_change_date":"2019-08-13T00:00:00Z","status_change_date":null,"inserted_date":null,"year_converted":0,"age":0,"original_entry_timestamp":null,"close_date":null,"cancellation_date":null,"pending_timestamp":null,"on_market_date":null,"contingent_date":null,"off_market_date":null,"cumulative_days_on_market":159,"modification_timestamp":null},"remarks":null,"marketing":{"virtual_tour_url_unbranded":"","internet_automated_valuation_display":false,"internet_consumer_comment":false,"internet_entire_listing_display":true,"internet_address_display":true,"website_detail_page_url":""},"closing":{"availability_date":null},"home_warranty":false},"tax":{"zoning":"RSA3","parcel_number":"541494285","tax_tract":"","tax_annual_amount":2345,"tax_year":2019,"tax_other_annual_assessment_amount":0},"hoa":{"association_fee":0,"association_fee_frequency":0,"association_fee2":0,"association_fee2_frequency":0,"association_amenities":"","association_fee_includes":"","pets_allowed":"False","association_name":""},"location":{"gis":{"cross_street":"","map_coordinate":"","directions":"South on Broad Street to right on 69th Street to left on 15th. Property on right.","latitude":40.05770492553711,"longitude":-75.14252471923828,"mls_latitude":0,"mls_longitude":0,"parcel_latitude":0,"parcel_longitude":0,"geocoded_city":"","neighborhood_id":""},"address":{"unparsed_address":"6822 N 15th Street N","city":"Philadelphia","county_or_parish":"Philadelphia","state_or_province":"PA","postal_code":"19126","street_dir_prefix":"","street_dir_suffix":"","street_name":"15TH","street_number":"6822","unit_number":"","township":""},"area":{"mls_area_major":"19126 (19126)","subdivision_name":""},"school":{"school_district":"The School District Of Philadelphia","elementary_school":"","elementary_school_district":"","middle_or_junior_school":"","high_school":"","high_school_district":"","middle_or_junior_school_district":"","sd_unif_id":"","sd_elem_id":"","sd_sec_id":"","saz_elem_id":"","saz_middle_id":"","saz_high_id":""}},"structure":{"architecture_style":"","heating":"Hot Water","cooling":"None","construction_materials":"","flooring":"Carpet,Hardwood","levels":"Main,Upper 1,Upper 2","bedrooms_total":5,"bathrooms_full":1,"bathrooms_partial":0,"bathrooms_onequarter":0,"bathrooms_threequarter":0,"bathrooms_half":0,"building_name":"","building_features":"","building_area_total":1705,"building_area_source":"","living_area":1705,"roof":"","parking_features":"","parking_total":"","other_parking":"","other_parking_spaces":"0","garage_spaces":0,"carport_spaces":0,"covered_spaces":0,"open_parking_spaces":0,"year_built":1925,"stories_total":0,"fireplace":false,"fireplace_features":"","fireplace_total":0,"door_features":"","foundation_details":"","insulation_desc":"","room_type":"","builder_name":"","accessibility_features":"","flood_area":"","homeowners_prot_plan":"","exterior_features":"","interior_features":"","body_type":"","basement":"Other","other_structures":"","builder_model":"","structure_type":"","rooms":{"rooms_total":8,"kitchen_dim":"","living_rm_dim":"","master_br_dim":"","dining_rm_dim":"","family_rm_dim":"","dining_desc":"Dining Room","family_room_desc":"","kitchen_desc":"Kitchen","living_rm_desc":"Living Room","bedroom_desc":"Master Bedroom","bathroom_desc":""},"property_condition":{"is_fixer_upper":false,"is_new_construction":false},"approximate_office_squarefeet":0,"approximate_retail_squarefeet":0,"approximate_warehouse_squarefeet":0,"total_restrooms":0,"load":""},"characteristics":{"lot_size_acres":"0.04","lot_size_dimensions":"25.00 x 72.75","lot_features":"","lot_size_square_feet":1819,"pool_features":"","private_pool":false,"view":"","laundry_features":"Basement","spa_features":"","community_features":"","complex_name":"","number_of_units_in_community":0,"water_body_name":"","water_front_features":"","water_front":false,"frontage_type":"","number_of_units_total":0,"hide_from_prelogin_search":false,"senior_community":false,"is_smart_home":false,"current_use":"","possible_use":"","number_of_lots":0,"number_of_pads":0,"development_status":"","fencing":"","road_surface_type":"","road_responsibility":"","misc_utilities_desc":"","furnished":"False","leaseterm":"","is_renters_insurance_required":false},"utilities":{"water_source":"Public","sewer":"Public Sewer","utilities":"","number_of_separate_electricmeters":0,"number_of_separate_gasmeters":0,"number_of_separate_watermeters":0},"equipment":{"other_equipment":"","appliances":"","security_features":""},"business":{"ownership_type":"Fee Simple","lease_amount_frequency":""}},"media":{"num_images":0,"modification_timestamp":null,"image_hash_code":"","media_info":[{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/B5B2a986E290482/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/b3F0BB492755445/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/B31B253B5ceb495/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/b070E301C868437/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/942072854b57419/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/a2f2c2310241421/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/19AE22156613476/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/0dbE64BCd60941f/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/CeB19B33EFe744d/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":416,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/92757AD75f6e48c/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/07294bDBa407432/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":519,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/Bf30F73BC452492/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/0aF2EF59cC1E47c/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/86c22c6Aa8c4444/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/9F59eb7B93BE42E/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""}]},"open_house":null,"dash":{"listing_guid":"98765","source_system_key":"","listing_agent_guid":"87654","company_staff_guid":""},"master_id":{"listing_master_id":"Q00800000FBSuCmkxVQDrAkzMxXsZjA4CaPz1peL","property_master_id":"","list_agent_mls_master_id":"3142663","list_office_mls_master_id":"","co_list_agent_mls_master_id":"","co_list_office_mls_master_id":"","buyer_agent_mls_master_id":"","buyer_office_mls_master_id":"","co_buyer_agent_mls_master_id":"","co_buyer_office_mls_master_id":"","address_master_id":"Q01000000FBSsXld8TMmLrafGFGGIsSHS9P342Yh"}}]}`)

	// This json is an array of listing100018 and listing29071815WithDash
	listing100018And29071815 = string(`{"mls_listings":[{"property":{"property_type":"Single Family","property_sub_type":"Twin/Semi-Detached","photos_count":15,"financial":{"rent_includes":"","sales_includes":"","electric_expense":0,"tenant_pays":"","owner_pays":"","income_includes":"","is_rent_control":false,"total_actual_rent":0},"listing":{"listing_id":"29071815","source_system_key":"BRIGHTMLS","source_system_name":"","standard_status":"ACTIVE","building_permits":"","documents_available":"","disclosures":"","contract":{"current_financing":"","special_listing_conditions":{"is_foreclosure":false,"is_short_sale":false,"is_probate_sale":false}},"price":{"list_price":115000,"list_price_high":0,"is_price_reduced":false,"price_change_timestamp":null,"original_list_price":0,"close_price":0,"pet_rent":0,"months_rent_upfront":0},"agent_office":{"list_agent":{"list_agent_fullname":"Anne Fitzgerald","list_agent_mls_id":"3172336","list_agent_office_phone":"2159387800","list_agent_office_phone_type":"","list_agent_state_license":"","list_agent_state_license_state":"","list_agent_email":"","list_agent_active":false,"list_agent_address":"","list_agent_city":"","list_agent_state_or_province":"","list_agent_postal_code":"","list_agent_type":"","list_agent_original_entry_timestamp":null,"list_agent_modification_timestamp":null},"list_office":{"list_office_name":"Better Homes Realty Group","list_office_phone":"","list_office_mls_id":"55916","list_office_address":"","list_office_city":"","list_office_state_or_province":"","list_office_postal_code":"","list_office_email":"","list_office_fax":"","list_office_original_entry_timestamp":null,"list_office_modification_timestamp":null},"co_list_agent":{"co_list_agent_full_name":"","co_list_agent_mls_id":"","co_list_agent_office_phone":""},"co_list_office":{"co_list_office_name":"","co_list_office_mls_id":"","co_list_office_phone":""},"buyer_agent":{"buyer_agent_fullname":"","buyer_agent_mls_id":"","buyer_office_phone":""},"buyer_office":null,"co_buyer_agent":null,"co_buyer_office":null},"compensation":null,"dates":{"listing_contract_date":"2019-03-07T00:00:00Z","first_appeared_date":null,"expiration_date":null,"last_change_date":"2019-08-13T00:00:00Z","status_change_date":null,"inserted_date":null,"year_converted":0,"age":0,"original_entry_timestamp":null,"close_date":null,"cancellation_date":null,"pending_timestamp":null,"on_market_date":null,"contingent_date":null,"off_market_date":null,"cumulative_days_on_market":159,"modification_timestamp":null},"remarks":null,"marketing":{"virtual_tour_url_unbranded":"","internet_automated_valuation_display":false,"internet_consumer_comment":false,"internet_entire_listing_display":true,"internet_address_display":true,"website_detail_page_url":""},"closing":{"availability_date":null},"home_warranty":false},"tax":{"zoning":"RSA3","parcel_number":"541494285","tax_tract":"","tax_annual_amount":2345,"tax_year":2019,"tax_other_annual_assessment_amount":0},"hoa":{"association_fee":0,"association_fee_frequency":0,"association_fee2":0,"association_fee2_frequency":0,"association_amenities":"","association_fee_includes":"","pets_allowed":"False","association_name":""},"location":{"gis":{"cross_street":"","map_coordinate":"","directions":"South on Broad Street to right on 69th Street to left on 15th. Property on right.","latitude":40.05770492553711,"longitude":-75.14252471923828,"mls_latitude":0,"mls_longitude":0,"parcel_latitude":0,"parcel_longitude":0,"geocoded_city":"","neighborhood_id":""},"address":{"unparsed_address":"6822 N 15th Street N","city":"Philadelphia","county_or_parish":"Philadelphia","state_or_province":"PA","postal_code":"19126","street_dir_prefix":"","street_dir_suffix":"","street_name":"15TH","street_number":"6822","unit_number":"","township":""},"area":{"mls_area_major":"19126 (19126)","subdivision_name":""},"school":{"school_district":"The School District Of Philadelphia","elementary_school":"","elementary_school_district":"","middle_or_junior_school":"","high_school":"","high_school_district":"","middle_or_junior_school_district":"","sd_unif_id":"","sd_elem_id":"","sd_sec_id":"","saz_elem_id":"","saz_middle_id":"","saz_high_id":""}},"structure":{"architecture_style":"","heating":"Hot Water","cooling":"None","construction_materials":"","flooring":"Carpet,Hardwood","levels":"Main,Upper 1,Upper 2","bedrooms_total":5,"bathrooms_full":1,"bathrooms_partial":0,"bathrooms_onequarter":0,"bathrooms_threequarter":0,"bathrooms_half":0,"building_name":"","building_features":"","building_area_total":1705,"building_area_source":"","living_area":1705,"roof":"","parking_features":"","parking_total":"","other_parking":"","other_parking_spaces":"0","garage_spaces":0,"carport_spaces":0,"covered_spaces":0,"open_parking_spaces":0,"year_built":1925,"stories_total":0,"fireplace":false,"fireplace_features":"","fireplace_total":0,"door_features":"","foundation_details":"","insulation_desc":"","room_type":"","builder_name":"","accessibility_features":"","flood_area":"","homeowners_prot_plan":"","exterior_features":"","interior_features":"","body_type":"","basement":"Other","other_structures":"","builder_model":"","structure_type":"","rooms":{"rooms_total":8,"kitchen_dim":"","living_rm_dim":"","master_br_dim":"","dining_rm_dim":"","family_rm_dim":"","dining_desc":"Dining Room","family_room_desc":"","kitchen_desc":"Kitchen","living_rm_desc":"Living Room","bedroom_desc":"Master Bedroom","bathroom_desc":""},"property_condition":{"is_fixer_upper":false,"is_new_construction":false},"approximate_office_squarefeet":0,"approximate_retail_squarefeet":0,"approximate_warehouse_squarefeet":0,"total_restrooms":0,"load":""},"characteristics":{"lot_size_acres":"0.04","lot_size_dimensions":"25.00 x 72.75","lot_features":"","lot_size_square_feet":1819,"pool_features":"","private_pool":false,"view":"","laundry_features":"Basement","spa_features":"","community_features":"","complex_name":"","number_of_units_in_community":0,"water_body_name":"","water_front_features":"","water_front":false,"frontage_type":"","number_of_units_total":0,"hide_from_prelogin_search":false,"senior_community":false,"is_smart_home":false,"current_use":"","possible_use":"","number_of_lots":0,"number_of_pads":0,"development_status":"","fencing":"","road_surface_type":"","road_responsibility":"","misc_utilities_desc":"","furnished":"False","leaseterm":"","is_renters_insurance_required":false},"utilities":{"water_source":"Public","sewer":"Public Sewer","utilities":"","number_of_separate_electricmeters":0,"number_of_separate_gasmeters":0,"number_of_separate_watermeters":0},"equipment":{"other_equipment":"","appliances":"","security_features":""},"business":{"ownership_type":"Fee Simple","lease_amount_frequency":""}},"media":{"num_images":0,"modification_timestamp":null,"image_hash_code":"","media_info":[{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/B5B2a986E290482/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/b3F0BB492755445/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/B31B253B5ceb495/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/b070E301C868437/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/942072854b57419/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/a2f2c2310241421/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/19AE22156613476/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/0dbE64BCd60941f/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/CeB19B33EFe744d/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":416,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/92757AD75f6e48c/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/07294bDBa407432/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":519,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/Bf30F73BC452492/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/0aF2EF59cC1E47c/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/86c22c6Aa8c4444/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""},{"media_url":"https://m.cbhomes.com/p/974/PAPH724168/9F59eb7B93BE42E/original.jpg","photos_change_timestamp":null,"image_height":640,"image_width":480,"md5":""}]},"open_house":null,"dash":{"listing_guid":"98765","source_system_key":"","listing_agent_guid":"87654","company_staff_guid":""},"master_id":{"listing_master_id":"Q00800000FBSuCmkxVQDrAkzMxXsZjA4CaPz1peL","property_master_id":"","list_agent_mls_master_id":"3142663","list_office_mls_master_id":"","co_list_agent_mls_master_id":"","co_list_office_mls_master_id":"","buyer_agent_mls_master_id":"","buyer_office_mls_master_id":"","co_buyer_agent_mls_master_id":"","co_buyer_office_mls_master_id":"","address_master_id":"Q01000000FBSsXld8TMmLrafGFGGIsSHS9P342Yh"}},{"property":{"property_type":"LAND","property_sub_type":"","financial":null,"listing":{"listing_id":"100018","source_system_key":"KY_WKRMLS","source_system_name":"","standard_status":"ACTIVE","building_permits":"","documents_available":"","disclosures":"","contract":{"current_financing":"","special_listing_conditions":{"is_foreclosure":false,"is_short_sale":false,"is_probate_sale":false}},"price":{"list_price":15000,"list_price_high":0,"is_price_reduced":true,"price_change_timestamp":null,"original_list_price":19000,"close_price":0,"pet_rent":0,"months_rent_upfront":0},"agent_office":{"list_agent":{"list_agent_fullname":"JOEY RICKS","list_agent_mls_id":"21","list_agent_office_phone":"","list_agent_office_phone_type":"","list_agent_state_license":"","list_agent_state_license_state":"","list_agent_email":"","list_agent_active":false,"list_agent_address":"","list_agent_city":"","list_agent_state_or_province":"","list_agent_postal_code":"","list_agent_type":"","list_agent_original_entry_timestamp":null,"list_agent_modification_timestamp":null},"list_office":{"list_office_name":"COLDWELL BANKER SERVICE 1ST REALTY","list_office_phone":"800-522-4699","list_office_mls_id":"23","list_office_address":"","list_office_city":"","list_office_state_or_province":"","list_office_postal_code":"","list_office_email":"","list_office_fax":"","list_office_original_entry_timestamp":null,"list_office_modification_timestamp":null},"co_list_agent":null,"co_list_office":null,"buyer_agent":null,"buyer_office":null,"co_buyer_agent":null,"co_buyer_office":null},"compensation":{"list_agency_compensation":null,"buyer_agency_compensation":{"percentage":"","fee":""}},"dates":{"listing_contract_date":null,"first_appeared_date":null,"expiration_date":null,"last_change_date":null,"status_change_date":null,"inserted_date":null,"year_converted":0,"age":0,"original_entry_timestamp":null,"close_date":null,"cancellation_date":null,"pending_timestamp":null,"on_market_date":null,"contingent_date":null,"off_market_date":null,"cumulative_days_on_market":0,"modification_timestamp":null},"remarks":{"public_remarks":"Nice corner lot at Arrow Head Golf Course. This lot is ready to be built on.","private_remarks":"","misc_info":"","selling_comments":""},"marketing":{"virtual_tour_url_unbranded":"","internet_automated_valuation_display":true,"internet_consumer_comment":true,"internet_entire_listing_display":false,"internet_address_display":false,"website_detail_page_url":""},"closing":null,"home_warranty":false},"tax":{"zoning":"","parcel_number":"","tax_tract":"Cherokee Hil","tax_annual_amount":0,"tax_year":0,"tax_other_annual_assessment_amount":0},"hoa":{"association_fee":0,"association_fee_frequency":0,"association_fee2":0,"association_fee2_frequency":0,"association_amenities":"","association_fee_includes":"","pets_allowed":"","association_name":""},"location":{"gis":{"cross_street":"","map_coordinate":"","directions":"","latitude":36.8587035,"longitude":-87.8069197,"mls_latitude":36.8287,"mls_longitude":-87.87515708,"parcel_latitude":0,"parcel_longitude":0,"geocoded_city":"Cadiz","neighborhood_id":""},"address":{"unparsed_address":"000 Comanche","city":"Cadiz","county_or_parish":"Trigg","state_or_province":"KY","postal_code":"42211","street_dir_prefix":"","street_dir_suffix":"","street_name":"","street_number":"","unit_number":"","township":""},"area":{"mls_area_major":"","subdivision_name":""},"school":{"school_district":"","elementary_school":"","elementary_school_district":"","middle_or_junior_school":"Trigg County","high_school":"Trigg County","high_school_district":"","middle_or_junior_school_district":"","sd_unif_id":"2105580","sd_elem_id":"","sd_sec_id":"","saz_elem_id":"KY-TRIGG CO-5702-KY-PB25665","saz_middle_id":"KY-TRIGG CO-5702-KY-PB25663","saz_high_id":"KY-TRIGG CO-5702-KY-PB25664"}},"structure":{"architecture_style":"","heating":"","cooling":"","construction_materials":"","flooring":"","levels":"","bedrooms_total":0,"bathrooms_full":0,"bathrooms_partial":0,"bathrooms_onequarter":0,"bathrooms_threequarter":0,"bathrooms_half":0,"building_name":"","building_features":"","building_area_total":0,"building_area_source":"","living_area":0,"roof":"","parking_features":"","parking_total":"","other_parking":"","other_parking_spaces":"","garage_spaces":0,"carport_spaces":0,"covered_spaces":0,"open_parking_spaces":0,"year_built":0,"stories_total":0,"fireplace":false,"fireplace_features":"","fireplace_total":0,"door_features":"","foundation_details":"","insulation_desc":"","room_type":"","builder_name":"","accessibility_features":"","flood_area":"","homeowners_prot_plan":"","exterior_features":"","interior_features":"","body_type":"","basement":"","other_structures":"","builder_model":"","structure_type":"","rooms":{"rooms_total":0,"kitchen_dim":"","living_rm_dim":"","master_br_dim":"","dining_rm_dim":"","family_rm_dim":"","dining_desc":"","family_room_desc":"","kitchen_desc":"","living_rm_desc":"","bedroom_desc":"","bathroom_desc":""},"property_condition":{"is_fixer_upper":false,"is_new_construction":false},"approximate_office_squarefeet":0,"approximate_retail_squarefeet":0,"approximate_warehouse_squarefeet":0,"total_restrooms":0,"load":""},"characteristics":{"lot_size_acres":"","lot_size_dimensions":"","lot_features":"","lot_size_square_feet":8712,"pool_features":"","private_pool":false,"view":"","laundry_features":"","spa_features":"","community_features":"","complex_name":"","number_of_units_in_community":0,"water_body_name":"","water_front_features":"","water_front":false,"frontage_type":"","number_of_units_total":0,"hide_from_prelogin_search":false,"senior_community":false,"is_smart_home":false,"current_use":"","possible_use":"","number_of_lots":0,"number_of_pads":0,"development_status":"","fencing":"","road_surface_type":"","road_responsibility":"","misc_utilities_desc":"","furnished":"","leaseterm":"","is_renters_insurance_required":false},"utilities":{"water_source":"","sewer":"","utilities":"","number_of_separate_electricmeters":0,"number_of_separate_gasmeters":0,"number_of_separate_watermeters":0},"equipment":{"other_equipment":"","appliances":"","security_features":""},"business":{"ownership_type":"","lease_amount_frequency":""}},"media":null,"open_house":null,"dash":null,"master_id":null}]}`)
)

var mockClient *mock.MockMlsListingServiceClient

func TestGetMlsListingByListingId(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingByListingIdRequest{ListingId: "100018"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingByListingIdResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingByListingId(
		ctx,
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingByListingId(ctx, &pb.GetMlsListingByListingIdRequest{ListingId: "100018"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingByListingGuid(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingByListingGuidRequest{ListingGuid: "98765"}
	testListingJson := listing29071815WithDash
	testListings := &pb.GetMlsListingByListingGuidResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingByListingGuid(
		ctx,
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingByListingGuid(ctx, &pb.GetMlsListingByListingGuidRequest{ListingGuid: "98765"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingBySource(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsBySourceRequest{SourceSystemKey: "KY_WKRMLS"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingsBySourceResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingBySource(
		gomock.Any(),
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingBySource(ctx, &pb.GetMlsListingsBySourceRequest{SourceSystemKey: "KY_WKRMLS"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)

	requestCommaSeperated := &pb.GetMlsListingsBySourceRequest{SourceSystemKey: "KY_WKRMLS, BRIGHTMLS"}
	testListingsCommaSeparated := &pb.GetMlsListingsBySourceResponse{}
	testListingListJson := listing100018And29071815
	json.Unmarshal([]byte(testListingListJson), testListingsCommaSeparated)

	mockClient.EXPECT().GetMlsListingBySource(
		gomock.Any(),
		requestCommaSeperated,
	).Return(testListingsCommaSeparated, nil)

	// Test comma separated input with space
	response, err = mockClient.GetMlsListingBySource(ctx, &pb.GetMlsListingsBySourceRequest{SourceSystemKey: "KY_WKRMLS, BRIGHTMLS"})

	assert.Equal(t, 2, len(response.MlsListings))
	assert.ElementsMatch(t, testListingsCommaSeparated.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestAddListings(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.MlsListingInput{
		ListingId: "456778",
		Property: &pb.PropertyInput{
			PropertyType: "LAND",
			Listing: &pb.ListingInput{
				StandardStatus: "ACTIVE",
				Price: &pb.PriceInput{
					ListPrice: 60000,
				},
			},
			Location: &pb.LocationInput{
				Address: &pb.AddressInput{
					City:    "Irving",
					Country: "India",
				},
			},
		},
	}
	mockClient.EXPECT().AddMlsListings(
		ctx,
		request,
	).Return(nil, errors.New("rpc error: code = InvalidArgument desc = Invalid input. RdmSourceSystemKey: cannot be blank."))

	response, err := mockClient.AddMlsListings(ctx, request)

	expectedError := status.Errorf(codes.InvalidArgument, "Invalid input. RdmSourceSystemKey: cannot be blank.")

	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestUpdateListings(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.UpdateMlsListingByListingIdRequest{
		ListingId: "456778",
		Property: &pb.UpdateProperty{
			Listing: &pb.UpdateListing{
				StandardStatus: "SOLD",
			},
		},
	}
	mockClient.EXPECT().UpdateMlsListingByListingId(
		ctx,
		request,
	).Return(nil, errors.New("rpc error: code = InvalidArgument desc = Closed Date and Close Price required for SOLD listing and Close date must be later than ListingContractDate"))

	response, err := mockClient.UpdateMlsListingByListingId(ctx, request)

	expectedError := status.Errorf(codes.InvalidArgument, "Closed Date and Close Price required for SOLD listing and Close date must be later than ListingContractDate")

	assert.NotNil(t, err)
	assert.Nil(t, response)
	assert.Equal(t, err.Error(), expectedError.Error())

	request1 := &pb.UpdateMlsListingByListingIdRequest{
		ListingId: "456778",
		Property: &pb.UpdateProperty{
			Listing: &pb.UpdateListing{
				StandardStatus: "ACTIVE",
				Price:          &pb.UpdatePrice{ListPrice: 20000},
			},
		},
	}
	mockClient.EXPECT().UpdateMlsListingByListingId(
		ctx,
		request1,
	).Return(&pb.UpdateMlsListingByListingIdResponse{MlsListings: &pb.MlsListing{Property: &pb.Property{Listing: &pb.Listing{StandardStatus: "ACTIVE", Price: &pb.Price{ListPrice: 20000}}}}}, nil)

	response1, err := mockClient.UpdateMlsListingByListingId(ctx, request1)

	assert.Nil(t, err)
	assert.NotNil(t, response1)
	assert.Equal(t, request1.Property.Listing.Price.ListPrice, response1.MlsListings.Property.Listing.Price.ListPrice)
	assert.Equal(t, request1.Property.Listing.StandardStatus, response1.MlsListings.Property.Listing.StandardStatus)
}

func TestGetMlsListingsByCity(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsByCityRequest{City: "Cadiz", State: "KY"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingsByCityResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByCity(
		gomock.Any(),
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingsByCity(ctx, &pb.GetMlsListingsByCityRequest{City: "Cadiz", State: "KY"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingsByState(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsByStateRequest{State: "KY"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingsByStateResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByState(
		gomock.Any(),
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingsByState(ctx, &pb.GetMlsListingsByStateRequest{State: "KY"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingsByPostalCode(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsByPostalCodeRequest{PostalCode: "42211"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingsByPostalCodeResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByPostalCode(
		ctx,
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingsByPostalCode(ctx, &pb.GetMlsListingsByPostalCodeRequest{PostalCode: "42211"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingsByAgentId(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request1 := &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21"}
	request2 := &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21", SourceSystemKey: "KY_WKRMLS"}
	request3 := &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21", SourceSystemKey: "BRIGHTMLS"}
	testListingJson := listing100018
	testListings := &pb.GetMlsListingsByAgentIdResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByAgentId(
		gomock.Any(),
		request1,
	).Return(testListings, nil)

	mockClient.EXPECT().GetMlsListingsByAgentId(
		gomock.Any(),
		request2,
	).Return(testListings, nil)

	mockClient.EXPECT().GetMlsListingsByAgentId(
		gomock.Any(),
		request3,
	).Return(nil, status.Errorf(codes.NotFound, fmt.Sprintf("Unable to find mls listings for ...")))

	response, err := mockClient.GetMlsListingsByAgentId(ctx, &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)

	response, err = mockClient.GetMlsListingsByAgentId(ctx, &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21", SourceSystemKey: "KY_WKRMLS"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)

	response, err = mockClient.GetMlsListingsByAgentId(ctx, &pb.GetMlsListingsByAgentIdRequest{ListAgentMlsId: "21", SourceSystemKey: "BRIGHTMLS"})

	assert.Nil(t, response)
	assert.NotNil(t, err)
}

func TestGetMlsListingsByAgentMasterId(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsByAgentMasterIdRequest{ListAgentMasterId: "Q00200000FDEuUMkQYUEOKYzUDCFObN1wo69zJlf"}
	testListingJson := listing29071815WithDash
	testListings := &pb.GetMlsListingsByAgentMasterIdResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByAgentMasterId(
		gomock.Any(),
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingsByAgentMasterId(ctx, &pb.GetMlsListingsByAgentMasterIdRequest{ListAgentMasterId: "Q00200000FDEuUMkQYUEOKYzUDCFObN1wo69zJlf"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func TestGetMlsListingsByAgentGuid(t *testing.T) {

	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mock.NewMockMlsListingServiceClient(ctrl)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &pb.GetMlsListingsByAgentGuidRequest{ListingAgentGuid: "87654"}
	testListingJson := listing29071815WithDash
	testListings := &pb.GetMlsListingsByAgentGuidResponse{}
	json.Unmarshal([]byte(testListingJson), testListings)

	mockClient.EXPECT().GetMlsListingsByAgentGuid(
		gomock.Any(),
		request,
	).Return(testListings, nil)

	response, err := mockClient.GetMlsListingsByAgentGuid(ctx, &pb.GetMlsListingsByAgentGuidRequest{ListingAgentGuid: "87654"})

	assert.Equal(t, 1, len(response.MlsListings))
	assert.ElementsMatch(t, testListings.MlsListings, response.MlsListings)
	assert.Nil(t, err)
}

func BenchmarkServer_GetMlsListingByListingId(b *testing.B) {

	conn, err := grpc.Dial(fmt.Sprintf("localhost:9080"), grpc.WithInsecure())
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for n := 0; n < b.N; n++ {

		if err != nil {
			log.Fatalf("Unable to connect: %v", err)
		}
		client := pb.NewMlsListingServiceClient(conn)

		request := &pb.GetMlsListingByListingIdRequest{ListingId: "100"}
		client.GetMlsListingByListingId(ctx, request)
	}
}
