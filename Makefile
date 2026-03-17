.PHONY: fmt format lint .build_migration_image .env-prod .env-dev .env-test up-prod up-dev up-test

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