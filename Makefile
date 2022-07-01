CURRENT_DIR = $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

MIGRATIONS_DIR = $(CURRENT_DIR)/migrations
SCRIPTS_DIR = $(CURRENT_DIR)/scripts
SOURCE_DIR = $(CURRENT_DIR)/src

MIGRATE = $(shell which migrate)
# Checks if migrate exist.
.PHONY: migrate-check
migrate-check:
	$(call error-if-empty,$(MIGRATE),migrate)

MIGRATE_DIR = $(MIGRATIONS_DIR)
MIGRATE_EXT = sql
# Creates a new migration.
.PHONY: migration
migration: migrate-check
	@echo "+ $@"
	@$(MIGRATE) create \
		-dir $(MIGRATE_DIR) \
		-ext $(MIGRATE_EXT) \
		-seq \
		$(NAME)

GOLANGCI_LINT = $(shell which golangci-lint)
# Checks if golangci-lint exist.
.PHONY: golangci-lint-check
golangci-lint-check:
	$(call error-if-emply,$(GOLANGCI_LINT),golangci-lint)

# Check lint, code styling rules. e.g. pylint, phpcs, eslint, style (java) etc ...
.PHONY: style
style:
	@echo "+ $@"
	@$(GOLANGCI_LINT) run \
		-v \
		"$(SCRIPTS_DIR)/..."

DOCKER_DIR = $(SCRIPTS_DIR)/docker

HADOLINT = $(shell which hadolint)
# Checks if hadolint exists.
.PHONY: hadolint-check
hadolint-check:
	$(call error-if-empty,$(HADOLINT),hadolint)

# Checks docker image with the server application.
.PHONY: hadolint-server
hadolint-server: hadolint-check
	@echo "+ $@"
	@$(HADOLINT) $(DOCKER_DIR)/server/Dockerfile

# Checks docker image with the schema.
.PHONY: hadolint-schema
hadolint-schema: hadolint-check
	@echo "+ $@"
	@$(HADOLINT) $(DOCKER_DIR)/schema/Dockerfile

# Checks docker image with the swagger.
.PHONY: hadolint-swagger
hadolint-swagger: hadolint-check
	@echo "+ $@"
	@$(HADOLINT) $(DOCKER_DIR)/swagger/Dockerfile

DOCKER = $(shell which docker)
# Checks if docker exists.
docker-check:
	$(call error-if-empty,$(DOCKER),docker)

# Build docker images.
.PHONY: docker-build
docker-build: docker-build-server docker-build-schema docker-build-swagger

# Builds docker image for the server.
.PHONY: docker-build-server
docker-build-server: docker-check hadolint-server
	@echo "+ $@"
	@$(DOCKER) build \
		--rm \
		-t server:latest \
		-f $(DOCKER_DIR)/server/Dockerfile \
		.

# Builds docker image for the percona schema.
.PHONY: docker-build-schema
docker-build-schema: docker-check hadolint-schema
	@echo "+ $@"
	@$(DOCKER) build \
		--rm \
		-t server-schema:latest \
		-f $(DOCKER_DIR)/schema/Dockerfile \
		.

# Builds docker image for the swagger.
.PHONY: docker-build-swagger
docker-build-swagger: docker-check hadolint-swagger
	@echo "+ $@"
	@$(DOCKER) build \
		--rm \
		-t server-swagger:latest \
		-f $(DOCKER_DIR)/swagger/Dockerfile \
		.

# Runs docker image with the server application.
.PHONY: docker-run-server
docker-run-server: docker-check docker-build
	@echo "+ $@"

define error-if-empty
@if [[ -z $(1) ]]; then echo "$(2) not installed"; false; fi
endef
