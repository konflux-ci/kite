.PHONY: migrate seed status migration

# Apply pending migrations
migrate:
	atlas migrate apply --env local

# Seed database (non-prod DBs)
seed:
	go run -mod=mod cmd/seed/main.go

# Get status of DB migrations (applied, pending)
status:
	atlas migrate status --env local

# Generate a new migration (requires NAME variable)
# Usage: make migration NAME="add_some_column"
migration:
	atlas migrate diff "$(NAME)" --env local
