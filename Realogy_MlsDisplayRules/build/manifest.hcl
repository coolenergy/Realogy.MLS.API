project {
  repository = "bitbucket.org/realogy_corp/mls-display-rules"
}

network {
  vpc_id = "${vpcId}"
}

variable "vpcId" {
  default = "vpc-0197252c503c62293"
  dev = "vpc-001c07632567a5d41"
  qa  = "vpc-001c07632567a5d41"
  uat = "vpc-0a56cddefc1e7c88e"
  prd = "vpc-0a56cddefc1e7c88e"
}

variable "env" {
  default = "default"
  dev = "dev"
  qa  = "qa"
  uat = "uat"
  prd = "prd"
}

variable "serverDesiredCount" {
  default = 2
  dev = 2
  qa  = 2
  uat = 4
  prd = 4
}

variable "serverCpu" {
  default = 256
  dev = 256
  qa = 256
  uat = 1024
  prd = 1024
}

variable "serverMemory" {
  default = 512
  dev = 512
  qa = 512
  uat = 2048
  prd = 2048
}

load_balancer "realogy-mls-display-rules" {
  hostname = "realogy-mls-display-rules.${hostname}"
}

variable "hostname" {
  default = ""
  dev = "dev.eapdenonprd.realogydev.com"
  qa = "qa.eapdenonprd.realogydev.com"
  uat = "uat.eapdeprd.realogyprod.com"
  prd = "cloud.realogyprod.com"
}

labels {
  business = "zap"
  product = "mls"
  application = "mls-display-rules"
  teamid = "mls"
  contact = "sakthi.palanisamy@realogy.com"
  classification = "unknown"
  compliance = "unknown"
}

task "server" {
  desired_count = "${serverDesiredCount}"
  cpu = "${serverCpu}"
  memory = "${serverMemory}"
  role = "${env}-mls-display-rules--role"
  env {
    HOST = "127.0.0.1"
    PORT = "80"
    GRPC_PORT = "9981"
    ENV = "${env}"
    NO_AUZ = true
  }
  target_group {
    load_balancer_name = "realogy-mls-display-rules"
  }
}

