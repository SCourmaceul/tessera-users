# Backlog

Points connus, non bloquants, à traiter plus tard.

## Dépendances et infrastructure

- [ ] **Version publiée de `tetra-kit`.** `go.mod` utilise `replace … => ../tetra-kit` : à remplacer par une version taguée une fois `tetra-kit` fusionné (et `GOPRIVATE=github.com/SCourmaceul` pour le dépôt privé), sinon la CI ne peut pas construire le service.
- [ ] **Topic SNS `tetra-events`.** Aucun repo ne le crée encore ; `events_topic_arn` suppose son ARN LocalStack.
- [ ] **Brancher le trigger Cognito.** Le user pool (Terraform de `tetra-gateway`) doit déclarer `lambda_config.post_confirmation` avec l'output `post_confirmation_lambda_arn`, et ce repo recevoir `cognito_user_pool_arn` pour la permission d'appel.

## Fonctionnel

- [ ] **Réputation.** Consommer `review.created` (file SQS abonnée à `tetra-events`) pour mettre à jour `reputation` du membre ; le format de l'événement est à définir avec `tetra-reviews`.
- [ ] **Unicité du pseudo.** Rien n'empêche deux utilisateurs d'avoir le même pseudo. Piste : un item `PSEUDO#<pseudo minuscule>` écrit dans une transaction avec le profil (nécessite `TransactWriteItems` dans `tetra-kit/adapter/dynamo`).
- [ ] **Rôle par espace.** Tout membre arrive avec `member` ; aucun handler ne le modifie encore (promotion `organizer`, à articuler avec les groupes Cognito).
- [ ] **Quitter un espace.**

## Robustesse

- [ ] **Outbox pour `user.created`.** Si la publication SNS échoue après la création du profil, l'événement est perdu (seulement journalisé).
- [ ] **Écritures non transactionnelles.** `update-me` écrit le profil puis chaque membre, `join-space` le membre puis le profil : un échec au milieu laisse une copie périmée (`join-space` se répare au prochain appel). Deux requêtes concurrentes sur le même profil peuvent aussi s'écraser (pas de verrou optimiste).
- [ ] **Pagination de `list-users`.** La partition d'un espace est lue entièrement ; ajouter `limit` / curseur avant que les espaces grossissent.
