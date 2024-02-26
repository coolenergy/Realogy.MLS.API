project {
  repository = "bitbucket.org/realogy_corp/mls-listings-service"
}

network {
  vpc_id = "${vpcId}"
}

variable "vpcId" {
  default = "vpc-0197252c503c62293"
  dev = "vpc-001c07632567a5d41"
  qa = "vpc-001c07632567a5d41"
  uat = "vpc-0a56cddefc1e7c88e"
  prd = "vpc-0a56cddefc1e7c88e"
}

load_balancer "realogy-mls-listings-service" {
  hostname = "realogy-mls-listings-service.${hostname}"
}

variable "hostname" {
  default = ""
  dev = "dev.eapdenonprd.realogydev.com"
  qa = "qa.eapdenonprd.realogydev.com"
  uat = "uat.eapdeprd.realogyprod.com"
  prd = "cloud.realogyprod.com"
}

task "server" {
  desired_count = "${serverDesiredCount}"
  cpu = "${serverCpu}"
  memory = "${serverMemory}"
  role = "${env}-mls-listings-service--role"
  env {
    GO_MLS_GATEWAY_PORT=80
    GO_MLS_MONGODB_PREFIX="mongodb+srv"
    GO_MLS_MONGODB_OPTIONS="${mongodbOptions}"
    GO_MLS_MONGODB_COLLECTIONS_LISTINGS="${mongodbCollectionListings}"
    GO_MLS_API_PAGINATION_LIMIT_DEFAULT=20
    GO_MLS_API_PAGINATION_LIMIT_MAX=250
    GO_MLS_API_BY_SOURCE_ALLOWED_LAST_CHANGE_DAYS=30
    GO_MLS_TRACING=true
    GO_MLS_AWS_SSM_BASEPATH="${awsSsmBasepath}"
    GO_MLS_AWS_LOCAL_ACTIVE="false"
    GO_MLS_API_AUTH_ACCESS_RULES="${authAccessRules}"
  }
  target_group {
    load_balancer_name = "realogy-mls-listings-service"
  }
  health_check {
    path = "/internal/health"
  }
}

variable "env" {
  default = "default"
  dev = "dev"
  qa = "qa"
  uat = "uat"
  prd = "prd"
}

variable "serverDesiredCount" {
  default = 2
  dev = 2
  qa = 2
  uat = 4
  prd = 4
}

variable "serverCpu" {
  default = 256
  dev = 256
  qa = 256
  uat = 1024
  prd = 2048
}

variable "serverMemory" {
  default = 512
  dev = 512
  qa = 512
  uat = 2048
  prd = 8192
}

variable "mongodbOptions" {
  default = ""
  dev = ""
  qa = ""
  uat = ""
  prd = ""
}

variable "mongodbCollectionListings" {
  default = "listings"
  dev = "listings"
  qa = "listings"
  uat = "listings"
  prd = "listings"
}

variable "awsSsmBasepath" {
  default = ""
  dev = "/realogy/services/mls-listings-service/dev/"
  qa = "/realogy/services/mls-listings-service/qa/"
  uat = "/realogy/services/mls-listings-service/uat/"
  prd = "/realogy/services/mls-listings-service/"
}

variable "authAccessRules"{
  default = ""
  dev = "0oaor7ejybgrubkqt0h7,[\"/realogy.api.mls.v1.MlsListingService/GetRealogyListings\"*\"/realogy.api.mls.v1.MlsListingService/GetMlsListingByListingId\"];0oa175di9npgjCepN0h8,[\"/realogy.api.mls.v1.MlsListingService/GetMlsListingByListingId\"]"
  qa  = ""
  uat = "0oa16g6vzl1Or2wwq0h8,[\"/realogy.api.mls.v1.MlsListingService/GetRealogyListings\"]"
  prd = ""
}

labels {
  business = "crp"
  product = "mls"
  application = "mls-listings-service"
  teamid = "eap"
  contact = "sakthi.palanisamy@zaplabs.com"
  classification = "unknown"
  compliance = "unknown"
}