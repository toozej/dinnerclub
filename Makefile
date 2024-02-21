# Set sane defaults for Make
SHELL = bash
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# Set default goal such that `make` runs `make help`
.DEFAULT_GOAL := help

# Build info
BUILDER = $(shell whoami)@$(shell hostname)
NOW = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Version control
VERSION = $(shell git describe --tags --dirty --always)
COMMIT = $(shell git rev-parse --short HEAD)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

# Linker flags
PKG = $(shell head -n 1 go.mod | cut -c 8-)
VER = $(PKG)/pkg/version
LDFLAGS = -s -w \
	-X $(VER).Version=$(or $(VERSION),unknown) \
	-X $(VER).Commit=$(or $(COMMIT),unknown) \
	-X $(VER).Branch=$(or $(BRANCH),unknown) \
	-X $(VER).BuiltAt=$(NOW) \
	-X $(VER).Builder=$(BUILDER)
	
OS = $(shell uname -s)
ifeq ($(OS), Linux)
	OPENER=xdg-open
else
	OPENER=open
endif

DEPLOY_HOSTNAME = $(shell grep DEPLOY_HOSTNAME ./deploy.env | awk -F= '{print $$2}')
DEPLOY_APPNAME = $(subst .,,$(DEPLOY_HOSTNAME))
DEPLOY_REGION = $(shell grep DEPLOY_REGION ./deploy.env | awk -F= '{print $$2}')
DEPLOY_POSTGRES_PASSWORD = $(shell grep DEPLOY_POSTGRES_PASSWORD ./deploy.env | awk -F= '{print $$2}')

.PHONY: all vet test build verify run up down distroless-build distroless-run local local-vet local-vendor local-test local-cover local-run local-release-test local-release local-sign local-verify local-release-verify install deploy deploy-secrets deploy-only deploy-ip deploy-cert deploy-psql-console deploy-backup deploy-restore get-cosign-pub-key docker-login pre-commit-install pre-commit-run pre-commit pre-reqs update-golang-version docs docs-generate docs-serve clean help

all: vet pre-commit clean test build verify run ## Run default workflow via Docker
local: local-update-deps local-vendor local-vet pre-commit clean local-test local-cover local-build local-sign local-verify local-run ## Run default workflow using locally installed Golang toolchain
local-release-verify: local-release local-sign local-verify ## Release and verify using locally installed Golang toolchain
pre-reqs: pre-commit-install ## Install pre-commit hooks and necessary binaries

vet: ## Run `go vet` in Docker
	docker build --target vet -f $(CURDIR)/Dockerfile -t toozej/dinnerclub:latest . 

test: ## Run `go test` in Docker
	docker build --target test -f $(CURDIR)/Dockerfile -t toozej/dinnerclub:latest . 
	@echo -e "\nStatements missing coverage"
	@grep -v -e " 1$$" c.out

build: ## Build Docker image, including running tests
	docker build -f $(CURDIR)/Dockerfile -t toozej/dinnerclub:latest .

get-cosign-pub-key: ## Get dinnerclub Cosign public key from GitHub
	test -f $(CURDIR)/dinnerclub.pub || curl --silent https://raw.githubusercontent.com/toozej/dinnerclub/main/dinnerclub.pub -O

verify: get-cosign-pub-key ## Verify Docker image with Cosign
	cosign verify --key $(CURDIR)/dinnerclub.pub toozej/dinnerclub:latest

run: ## Run built Docker image
	-docker kill dinnerclub
	docker run --rm -d --name dinnerclub -p 8080:8080 -v $(CURDIR)/config:/config toozej/dinnerclub:latest

up: ## Run Docker Compose project
	docker compose -f docker-compose.yml down --remove-orphans
	docker compose -f docker-compose.yml build --pull
	docker compose -f docker-compose.yml up -d

down: ## Stop running Docker Compose project
	docker compose -f docker-compose.yml down --remove-orphans

rebuild: ## Rebuild running application container in Docker Compose project
	docker compose -f docker-compose.yml stop dinnerclub_app
	docker compose -f docker-compose.yml build --pull dinnerclub_app
	docker compose -f docker-compose.yml up -d dinnerclub_app

logs: ## View running Docker Compose project logs
	docker compose -f docker-compose.yml logs --tail=100 -f

distroless-build: ## Build Docker image using distroless as final base
	docker build -f $(CURDIR)/Dockerfile.distroless -t toozej/dinnerclub:distroless . 

distroless-run: ## Run built Docker image using distroless as final base
	docker run --rm --name dinnerclub -v $(CURDIR)/config:/config toozej/dinnerclub:distroless

local-update-deps: ## Run `go get -t -u ./...` to update Go module dependencies
	go get -t -u ./...

local-vet: ## Run `go vet` using locally installed golang toolchain
	go vet $(CURDIR)/...

local-vendor: ## Run `go mod vendor` using locally installed golang toolchain
	go mod tidy
	go mod vendor

local-test: ## Run `go test` using locally installed golang toolchain
	go test -coverprofile c.out -v $(CURDIR)/...
	@echo -e "\nStatements missing coverage"
	@grep -v -e " 1$$" c.out

local-cover: ## View coverage report in web browser
	go tool cover -html=c.out

local-build: ## Run `go build` using locally installed golang toolchain
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" $(CURDIR)/cmd/dinnerclub/

local-run: ## Run locally built binary
	$(CURDIR)/dinnerclub

local-release-test: ## Build assets and test goreleaser config using locally installed golang toolchain and goreleaser
	goreleaser check
	goreleaser build --rm-dist --snapshot

local-release: local-test docker-login ## Release assets using locally installed golang toolchain and goreleaser
	if test -e $(CURDIR)/dinnerclub.key && test -e $(CURDIR)/cicd.env; then \
		export `cat $(CURDIR)/cicd.env | xargs` && git push origin `git describe --tags` && goreleaser release --rm-dist; \
	else \
		echo "Missing required cosign private key or environment variables. Cannot release."; \
	fi

local-sign: local-test ## Sign locally installed golang toolchain and cosign
	if test -e $(CURDIR)/dinnerclub.key && test -e $(CURDIR)/cicd.env; then \
		export `cat $(CURDIR)/cicd.env | xargs` && cosign sign-blob --key=$(CURDIR)/dinnerclub.key --output-signature=$(CURDIR)/dinnerclub.sig $(CURDIR)/dinnerclub; \
	else \
		echo "no cosign private key found at $(CURDIR)/dinnerclub.key. Cannot release."; \
	fi

local-verify: get-cosign-pub-key ## Verify locally compiled binary
	# cosign here assumes you're using Linux AMD64 binary
	cosign verify-blob --key $(CURDIR)/dinnerclub.pub --signature $(CURDIR)/dinnerclub.sig $(CURDIR)/dinnerclub

install: local-build local-verify ## Install compiled binary to local machine
	sudo cp $(CURDIR)/dinnerclub /usr/local/bin/dinnerclub
	sudo chmod 0755 /usr/local/bin/dinnerclub

deploy: deploy-secrets deploy-only deploy-ip deploy-cert ## Deploy to fly.io
	flyctl status

deploy-only: ## Deploy locally built runtime image to fly.io
	flyctl deploy $(CURDIR) --local-only

deploy-ip: ## Allocate an IP address for deployment in fly.io
	flyctl ips list | grep v4 || flyctl ips allocate-v4
	flyctl ips list

deploy-cert: ## Provision a SSL certificate for deployment in fly.io
	flyctl certs list | grep $(DEPLOY_HOSTNAME) || flyctl certs create "$(DEPLOY_HOSTNAME)"
	flyctl certs list

deploy-secrets: ## Deploy secrets to fly.io
	@if test -e $(CURDIR)/app.env; then \
		echo "Trying to load app secrets from app.env file"; \
		while read -r SECRET; do \
			if [[ "$${SECRET}" =~ .*SECRET.*|.*PASSWORD.*|.*REFERRAL.* ]]; then \
				flyctl secrets set --stage $${SECRET}; \
			fi; \
		done < $(CURDIR)/app.env
	else \
		echo "Trying to load app secrets from environment"; \
		while read -r SECRET; do \
			if [[ "$${SECRET}" =~ .*SECRET.*|.*PASSWORD.*|.*REFERRAL.* ]]; then \
				flyctl secrets set --stage $${SECRET}; \
			fi; \
		done < <(env); \
	fi
	flyctl config env

deploy-first-time-psql-setup: ## Deploy Postgres database in fly.io
	# TODO test fly postgres create command to find correct settings for development / single VM
	fly postgres create --name $(DEPLOY_APPNAME)-db --region $(DEPLOY_REGION) --initial-cluster-size 1
	# TODO add remainder of steps from README

deploy-psql-console: ## Run a PSQL console to the deployed dinnerclub database in fly.io
	flyctl proxy 5434:5432 -a $(DEPLOY_APPNAME)-db &
	sleep 5
	PGPASSWORD=$(DEPLOY_POSTGRES_PASSWORD) psql -h localhost -p 5434 -U $(DEPLOY_APPNAME) $(DEPLOY_APPNAME)
	pkill -15 -f 'flyctl proxy'

deploy-backup: ## Backup dinnerclub database in fly.io to localhost
	# https://fly.io/docs/postgres/managing/backup-and-restore/
	flyctl proxy 5434:5432 -a $(DEPLOY_APPNAME)-db &
	sleep 5
	PGPASSWORD=$(DEPLOY_POSTGRES_PASSWORD) pg_dump -h localhost -p 5434 -U $(DEPLOY_APPNAME) $(DEPLOY_APPNAME) > ./backups/flyio_$(DEPLOY_APPNAME)_dinnerclub_$(NOW).sql
	pkill -15 -f 'flyctl proxy'

deploy-restore: ## Restore dinnerclub database from localhost backup to fly.io
	# https://fly.io/docs/postgres/managing/backup-and-restore/
	flyctl proxy 5434:5432 -a $(DEPLOY_APPNAME)-db &
	sleep 5
	# TODO write and test restore command, finding most recent backup and restoring that
	PGPASSWORD=$(DEPLOY_POSTGRES_PASSWORD) pg_restore -h localhost -p 5434 -U $(DEPLOY_APPNAME) $(DEPLOY_APPNAME) < ./backups/flyio_$(DEPLOY_APPNAME)_dinnerclub_$(NOW).sql
	pkill -15 -f 'flyctl proxy'
	

docker-login: ## Login to Docker registries used to publish images to
	if test -e $(CURDIR)/cicd.env; then \
		export `cat $(CURDIR)/cicd.env | xargs`; \
		echo $${DOCKERHUB_TOKEN} | docker login docker.io --username $${DOCKERHUB_USERNAME} --password-stdin; \
		echo $${QUAY_TOKEN} | docker login quay.io --username $${QUAY_USERNAME} --password-stdin; \
		echo $${GITHUB_GHCR_TOKEN} | docker login ghcr.io --username $${GITHUB_USERNAME} --password-stdin; \
	else \
		echo "No container registry credentials found, need to add them to ./cicd.env. See README.md for more info"; \
	fi

pre-commit: pre-commit-install pre-commit-run ## Install and run pre-commit hooks

pre-commit-install: ## Install pre-commit hooks and necessary binaries
	# golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# goimports
	go install golang.org/x/tools/cmd/goimports@latest
	# gosec
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	# staticcheck
	go install honnef.co/go/tools/cmd/staticcheck@latest
	# go-critic
	go install github.com/go-critic/go-critic/cmd/gocritic@latest
	# structslop
	go install github.com/orijtech/structslop/cmd/structslop@latest
	# shellcheck
	command -v shellcheck || sudo dnf install -y ShellCheck || sudo apt install -y shellcheck
	# checkmake
	go install github.com/mrtazz/checkmake/cmd/checkmake@latest
	# goreleaser
	go install github.com/goreleaser/goreleaser@latest
	# syft
	command -v syft || curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
	# cosign
	go install github.com/sigstore/cosign/cmd/cosign@latest
	# go-licenses
	go install github.com/google/go-licenses@latest
	# go vuln check
	go install golang.org/x/vuln/cmd/govulncheck@latest
	# install and update pre-commits
	pre-commit install
	pre-commit autoupdate

pre-commit-run: ## Run pre-commit hooks against all files
	pre-commit run --all-files
	# manually run the following checks since their pre-commits aren't working or don't exist
	go-licenses report github.com/toozej/dinnerclub/cmd/dinnerclub
	govulncheck ./...

update-golang-version: ## Update to latest Golang version across the repo
	@VERSION=`curl -s "https://go.dev/dl/?mode=json" | jq -r '.[0].version' | sed 's/go//' | cut -d '.' -f 1,2`; \
	echo "Updating Golang to $$VERSION"; \
	./scripts/update_golang_version.sh $$VERSION

docs: docs-generate docs-serve ## Generate and serve documentation

docs-generate:
	docker build -f $(CURDIR)/Dockerfile.docs -t toozej/dinnerclub:docs . 
	docker run --rm --name dinnerclub-docs -v $(CURDIR):/package -v $(CURDIR)/docs:/docs toozej/dinnerclub:docs

docs-serve: ## Serve documentation on http://localhost:9000
	docker run -d --rm --name dinnerclub-docs-serve -p 9000:3080 -v $(CURDIR)/docs:/data thomsch98/markserv
	$(OPENER) http://localhost:9000/docs.md
	@echo -e "to stop docs container, run:\n"
	@echo "docker kill dinnerclub-docs-serve"

clean: ## Remove any locally compiled binaries
	rm -f $(CURDIR)/dinnerclub

help: ## Display help text
	@grep -E '^[a-zA-Z_-]+ ?:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
