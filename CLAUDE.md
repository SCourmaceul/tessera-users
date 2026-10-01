# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`tessera-users` is the **users** microservice of **Tessera** (formerly Tetra; the platform is being renamed tetra → tessera). This service's own names use `tessera`: Go module `github.com/SCourmaceul/tessera-users`, table `tessera-users`, Lambdas `tessera-users-<handler>`, event source `tessera-users`, registry `service_name = service#users`. Names owned by repos not yet renamed keep `tetra` until those repos change: `tetra-kit`, `tetra-gateway`, the `tetra-gateway-routes` registry, the `tetra-events` topic, the `X-Tetra-Space` header. It owns the public profile (pseudo, avatar), the spaces a user joined, and the per-space role and reputation. One Lambda per handler, invoked by `tetra-gateway`; the service catalogue lives in `../doc/tetra_services.md`. Code comments and API error messages are in French; error **codes** stay in plain English (`USER_NOT_FOUND`…). Don't reintroduce the old cyberpunk "HUD // SQUAD" vocabulary.

## Commands

- `make test` — `go test -v -race -count=1 ./...` (`-count=1` matters: `architecture_test.go` shells out to `go list`)
- `make build` — one `build/<handler>/bootstrap` per handler (linux/arm64, `provided.al2023`); the `HANDLERS` list in the Makefile and `local.handlers` in `terraform/main.tf` must stay in sync with `cmd/`
- `make tf-init`, `make deploy ADMIN_TOKEN=…` — Terraform against LocalStack, then `POST /admin/registry/reload` on the gateway

## Architecture

- Layout follows the `tetra-kit` README: `cmd/<handler>` wires, `internal/domain` (entities, validation, `errs.Error` values), `internal/app` (use cases + `Repository` port), `internal/adapter/{dynamo,http,cognito}`, `internal/platform` (AWS config from env: `TABLE_NAME`, `EVENTS_TOPIC_ARN`). `architecture_test.go` forbids `domain`/`app` from importing AWS, `net/http`, or any adapter.
- `tetra-kit` is pulled through `replace => ../tetra-kit` (private repo, unmerged); see BACKLOG.
- **Data model** (`internal/adapter/dynamo/store.go`): profile at `USER#<id>`/`PROFILE` (the only key not prefixed by a space — a profile belongs to none) and membership at `SPACE#<space>#MEMBERS`/`USER#<id>`, which carries a copy of pseudo/avatar so a space is listed in one query. `UpdateMe` must keep those copies in sync.
- **Space isolation:** a user is only visible (`get-user`, `list-users`) in spaces they joined; the space comes from `req.Space()` (trusted, set by the gateway).
- **Post-confirmation trigger** only acts on `PostConfirmation_ConfirmSignUp`, is idempotent (existing profile → no-op, no event) and must not fail the sign-up because of SNS: a `user.created` publish error is only logged. `user.created` is published without a space.
- Routes are declared in `terraform/main.tf` (`local.handlers`), all with `auth_required = true`. `/api/v1/users/me` is matched before `/:id` by the gateway trie (exact segments win).
