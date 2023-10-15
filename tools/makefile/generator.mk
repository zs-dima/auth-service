.PHONY: sql

sql: export SQLC_AUTH_TOKEN=test
sql:
	@echo "* Running sqlc generator *"
	@docker run --rm -v $(CURDIR):/src -w /src sqlc/sqlc generate --no-remote
