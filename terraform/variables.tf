variable "region" {
  type    = string
  default = "eu-west-3"
}

variable "localstack_endpoint" {
  type    = string
  default = "http://localhost:4566"
}

# Topic SNS partagé des événements Tessera (user.created est publié dessus)
variable "events_topic_arn" {
  type    = string
  default = "arn:aws:sns:eu-west-3:000000000000:tetra-events"
}

# Registre des routes, créé par tetra-gateway
variable "routes_table_name" {
  type    = string
  default = "tetra-gateway-routes"
}

# ARN du user pool Cognito (tetra-gateway, enable_cognito=true). Vide : pas de permission
# d'appel du trigger post-confirmation (cas de LocalStack, sans Cognito).
variable "cognito_user_pool_arn" {
  type    = string
  default = ""
}
