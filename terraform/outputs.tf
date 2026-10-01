output "table_name" {
  value = aws_dynamodb_table.users.name
}

# À déclarer comme trigger post-confirmation du user pool (tetra-gateway)
output "post_confirmation_lambda_arn" {
  value = aws_lambda_function.handler["post-confirmation"].arn
}

output "routes" {
  value = { for name, r in local.routes : name => "${r.method} ${r.path}" }
}
