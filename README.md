# tessera-users

Service **users** de **Tessera** : profil public (pseudo, avatar), espaces rejoints, rôle et réputation par espace. Une Lambda Go par handler, appelée par `tetra-gateway` ; le code commun vient de [`tetra-kit`](../tetra-kit). Catalogue des services : `../doc/tetra_services.md`.

## Handlers

| Lambda | Déclencheur | Rôle |
|---|---|---|
| `tessera-users-post-confirmation` | Trigger Cognito post-confirmation | Crée le profil (pseudo proposé depuis `preferred_username` ou l'e-mail) et publie `user.created` |
| `tessera-users-get-me` | `GET /api/v1/users/me` | Profil de l'utilisateur connecté et ses espaces |
| `tessera-users-update-me` | `PATCH /api/v1/users/me` | Modifie `pseudo` et/ou `avatar_url` |
| `tessera-users-join-space` | `POST /api/v1/users/me/spaces` | Rejoint l'espace de la requête (`X-Tetra-Space`) avec le rôle `member` |
| `tessera-users-get-user` | `GET /api/v1/users/:id` | Profil d'un membre de l'espace |
| `tessera-users-list-users` | `GET /api/v1/users` | Membres de l'espace, triés par pseudo |

Toutes les routes exigent un utilisateur authentifié. Un utilisateur n'est visible que dans les espaces qu'il a rejoints (`404 USER_NOT_FOUND` ailleurs).

Exemples de réponses :

```jsonc
// GET /api/v1/users/me
{"id": "…", "pseudo": "ada", "avatar_url": "", "created_at": "…",
 "spaces": [{"space": "core", "role": "member", "reputation": 0, "joined_at": "…"}]}

// GET /api/v1/users/:id (dans l'espace core)
{"id": "…", "pseudo": "ada", "avatar_url": "", "space": "core", "role": "member", "reputation": 0, "joined_at": "…"}

// GET /api/v1/users
{"users": [ … ]}
```

Erreurs (format de la gateway `{"error": CODE, "message": …}`) : `USER_NOT_FOUND` (404), `ALREADY_MEMBER` (409), `INVALID_PSEUDO`, `INVALID_AVATAR_URL`, `INVALID_SPACE`, `INVALID_BODY` (400), `UNAUTHORIZED` (401).

## Données

Table DynamoDB `tessera-users` (`pk` / `sk`) :

| Item | `pk` | `sk` | Attributs |
|---|---|---|---|
| Profil | `USER#<id>` | `PROFILE` | `id`, `pseudo`, `avatar_url`, `spaces`, `created_at`, `updated_at` |
| Membre d'un espace | `SPACE#<espace>#MEMBERS` | `USER#<id>` | `user_id`, `space`, `pseudo`, `avatar_url`, `role`, `reputation`, `joined_at` |

Le profil est la seule clé non préfixée par un espace : il est commun à tous. Le membre porte une copie du pseudo et de l'avatar pour lister un espace en une requête ; `update-me` la tient à jour.

## Structure

```
cmd/<handler>/main.go         câblage d'une Lambda
internal/domain/              profil, membre, validation (pseudo, avatar)
internal/app/                 cas d'usage et port Repository
internal/adapter/dynamo/      Repository sur la table DynamoDB
internal/adapter/http/        handlers lambdahttp
internal/adapter/cognito/     trigger post-confirmation
internal/platform/            configuration AWS commune aux Lambdas
terraform/                    table, Lambdas, rôle IAM, routes du registre de la gateway
```

## Développement

```bash
make test        # tests (dont la règle de dépendance, architecture_test.go)
make build       # build/<handler>/bootstrap (linux/arm64, provided.al2023)
make tf-init
make deploy ADMIN_TOKEN=…   # terraform apply sur LocalStack + rechargement du registre de la gateway
```

Prérequis du déploiement local : LocalStack lancé, Terraform de `tetra-gateway` appliqué (table `tetra-gateway-routes`), gateway démarrée pour `reload-routes`. Le topic SNS `tetra-events` n'est encore créé par aucun repo : `events_topic_arn` pointe vers son ARN LocalStack attendu.

`tetra-kit` est résolu par une directive `replace` vers `../tetra-kit` (dépôt privé, pas encore de version publiée) : cloner les deux dépôts côte à côte. La CI (`.github/workflows/ci.yml`) fait de même et a besoin du secret `TETRA_KIT_TOKEN`, un jeton avec accès en lecture à `tetra-kit`.
