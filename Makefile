.PHONY: \
fmt \
format \
lint \
.build_migration_image \
.env-prod \
.env-dev \
.env-test \
up-prod \
up-dev \
up-test \
db-bench \
db-bench-fast \
usecase-bench \
usecase-bench-fast \
integration-test \
integration-test-fast \
e2e-test \
e2e-test-fast \
gen-mocks \
unit-test

fmt:
	go fmt ./...

format:
	golines --max-len=120 . -l
	golines --max-len=120 . -w

lint:
	golangci-lint run

.build_migration_image:
	docker build -t goose-migration -f Dockerfile.migration .

.env-prod:
	APP_MODE=prod bash ./environment.sh

.env-dev:
	APP_MODE=dev bash ./environment.sh

.env-test:
	APP_MODE=test bash ./environment.sh

up-prod: .env-prod .build_migration_image
	docker compose --profile prod up

up-dev: .env-dev .build_migration_image
	docker compose --profile dev up -d

up-test: .env-test .build_migration_image
	docker compose --profile test up -d

db-bench: up-test
	go test ./tests/integration/db -bench=. -benchmem -benchtime=5s -run=^$$

db-bench-fast:
	go test ./tests/integration/db -bench=. -benchmem -benchtime=5s -run=^$$

usecase-bench: up-test
	go test ./tests/integration/usecases -bench=. -benchmem -benchtime=5s -run=^$$

usecase-bench-fast:
	go test ./tests/integration/usecases -bench=. -benchmem -benchtime=5s -run=^$$

integration-test: up-test
	go test ./tests/integration/... -count=1

integration-test-fast:
	go test ./tests/integration/... -count=1

e2e-test: up-test
	go test ./tests/e2e/... -count=1

e2e-test-fast:
	go test ./tests/e2e/... -count=1

gen-mocks:
	go generate ./internal/usecases/...

unit-test:
	go test ./internal/usecases/tests/... -count=1