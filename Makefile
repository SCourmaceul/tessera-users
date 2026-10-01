.PHONY: test vet fmt tidy build tf-init tf-apply deploy reload-routes clean

HANDLERS := post-confirmation get-me update-me join-space get-user list-users
GATEWAY_URL ?= http://localhost:8080

test:
	@echo ">> Running test suite..."
	@go test -v -race -count=1 ./...

vet:
	@go vet ./...

fmt:
	@gofmt -l -w .

tidy:
	@echo ">> Updating Go dependencies..."
	@go mod tidy

# Un binaire bootstrap par handler, pour le runtime provided.al2023 (arm64)
build:
	@echo ">> Building Lambdas..."
	@for h in $(HANDLERS); do \
		GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -trimpath -ldflags="-s -w" \
			-o build/$$h/bootstrap ./cmd/$$h || exit 1; \
	done

tf-init:
	@echo ">> Initializing Terraform..."
	@cd terraform && terraform init

tf-apply: build
	@echo ">> Applying Terraform (LocalStack)..."
	@cd terraform && terraform apply -auto-approve

# Déploie puis recharge le registre de la gateway
deploy: tf-apply reload-routes

# make reload-routes ADMIN_TOKEN=… (même valeur que dans le .env de tetra-gateway)
reload-routes:
	@test -n "$(ADMIN_TOKEN)" || (echo "ADMIN_TOKEN manquant" && exit 1)
	@curl -fsS -X POST -H "Authorization: Bearer $(ADMIN_TOKEN)" $(GATEWAY_URL)/admin/registry/reload && echo

clean:
	@rm -rf build/
