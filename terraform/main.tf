terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    archive = {
      source = "hashicorp/archive"
    }
  }
}

provider "aws" {
  region     = var.region
  access_key = "test"
  secret_key = "test"

  # Identifiants factices de LocalStack : désactive les vérifications AWS
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  # Redirige les appels vers LocalStack
  endpoints {
    dynamodb = var.localstack_endpoint
    iam      = var.localstack_endpoint
    lambda   = var.localstack_endpoint
    sts      = var.localstack_endpoint
  }
}

locals {
  service = "tessera-users"

  # Une Lambda par handler. Les handlers avec une route sont inscrits dans le registre de la
  # gateway ; post-confirmation est un trigger Cognito.
  handlers = {
    "post-confirmation" = { route = null, publishes = true }
    "get-me"            = { route = { method = "GET", path = "/api/v1/users/me", timeout_ms = 2000 }, publishes = false }
    "update-me"         = { route = { method = "PATCH", path = "/api/v1/users/me", timeout_ms = 3000 }, publishes = false }
    "join-space"        = { route = { method = "POST", path = "/api/v1/users/me/spaces", timeout_ms = 3000 }, publishes = false }
    "get-user"          = { route = { method = "GET", path = "/api/v1/users/:id", timeout_ms = 2000 }, publishes = false }
    "list-users"        = { route = { method = "GET", path = "/api/v1/users", timeout_ms = 3000 }, publishes = false }
  }

  routes = { for name, h in local.handlers : name => h.route if h.route != null }

  tags = {
    Environment = "lab"
    Project     = "tessera"
    Service     = local.service
  }
}

# Table du service : profils (USER#<id> / PROFILE) et membres des espaces
# (SPACE#<espace>#MEMBERS / USER#<id>). Jamais lue par un autre service.
resource "aws_dynamodb_table" "users" {
  name         = local.service
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "pk"
  range_key    = "sk"

  attribute {
    name = "pk"
    type = "S"
  }

  attribute {
    name = "sk"
    type = "S"
  }

  tags = local.tags
}

# Rôle d'exécution commun aux Lambdas du service
resource "aws_iam_role" "lambda" {
  name = "${local.service}-lambda"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
  tags = local.tags
}

resource "aws_iam_role_policy" "lambda" {
  name = "${local.service}-lambda"
  role = aws_iam_role.lambda.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:DeleteItem", "dynamodb:Query"]
        Resource = aws_dynamodb_table.users.arn
      },
      {
        Effect   = "Allow"
        Action   = "sns:Publish"
        Resource = var.events_topic_arn
      },
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
        Resource = "*"
      },
    ]
  })
}

# Binaires construits par `make build` (build/<handler>/bootstrap)
data "archive_file" "handler" {
  for_each    = local.handlers
  type        = "zip"
  source_file = "${path.module}/../build/${each.key}/bootstrap"
  output_path = "${path.module}/../build/${each.key}.zip"
}

resource "aws_lambda_function" "handler" {
  for_each         = local.handlers
  function_name    = "${local.service}-${each.key}"
  role             = aws_iam_role.lambda.arn
  runtime          = "provided.al2023"
  architectures    = ["arm64"]
  handler          = "bootstrap"
  filename         = data.archive_file.handler[each.key].output_path
  source_code_hash = data.archive_file.handler[each.key].output_base64sha256
  memory_size      = 128
  timeout          = 10

  environment {
    variables = merge(
      { TABLE_NAME = aws_dynamodb_table.users.name },
      each.value.publishes ? { EVENTS_TOPIC_ARN = var.events_topic_arn } : {},
    )
  }

  tags = local.tags
}

# Autorise le user pool à appeler le trigger (le user pool, dans tetra-gateway, doit déclarer
# lambda_config.post_confirmation = output post_confirmation_lambda_arn)
resource "aws_lambda_permission" "cognito_post_confirmation" {
  count         = var.cognito_user_pool_arn == "" ? 0 : 1
  statement_id  = "AllowCognitoPostConfirmation"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.handler["post-confirmation"].function_name
  principal     = "cognito-idp.amazonaws.com"
  source_arn    = var.cognito_user_pool_arn
}

# Inscription des routes dans le registre de tetra-gateway. Après un apply, recharger le
# registre : make reload-routes
resource "aws_dynamodb_table_item" "route" {
  for_each   = local.routes
  table_name = var.routes_table_name
  hash_key   = "service_name"
  range_key  = "route_id"

  item = jsonencode({
    service_name  = { S = "service#users" }
    route_id      = { S = "${each.value.method}#${each.value.path}" }
    method        = { S = each.value.method }
    path_pattern  = { S = each.value.path }
    target_arn    = { S = aws_lambda_function.handler[each.key].arn }
    auth_required = { BOOL = true }
    is_active     = { BOOL = true }
    timeout_ms    = { N = tostring(each.value.timeout_ms) }
  })
}
