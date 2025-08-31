# Define the output for project's binary
OUTPUT = bin/fury-flow

# Default target
all: build

# Define the target: build
build:
	mkdir -p bin
	go build -o $(OUTPUT) ./cmd/fury-flow
	echo "Build completed: $(OUTPUT)"

# Define the target: build-goose (local goose)
# we want to use it in the context of this project only (no global installation)
build-goose:
	mkdir -p bin
	go get github.com/pressly/goose/v3/cmd/goose
	go build -o ./bin/goose github.com/pressly/goose/v3/cmd/goose
	echo "[GOOSE] Build completed"

# Define the target: goose-migrate
# use that locally installed/build goose application/tool
# this command expect env variables to be set: GOOSE_
goose-migrate:
	./bin/goose up

# Define the target: goose-migrate-reset
# use that locally installed/build goose application/tool
# this command expect env variables to be set: GOOSE_
goose-migrate-reset:
	./bin/goose reset

# Define the target: build-sqlc (local sqlc)
# we want to use it in the context of this project only (no global installation)
build-sqlc:
	mkdir -p bin
	go get github.com/sqlc-dev/sqlc/cmd/sqlc
	go build -o ./bin/sqlc github.com/sqlc-dev/sqlc/cmd/sqlc
	echo "[SQLC] Build completed"

# Define the target: sqlc-generate
sqlc-generate:
	./bin/sqlc generate

# Define the target: run
run: build
	$(OUTPUT)

# Define the target: test-unit
test-unit:
	go test -v -tags=unit ./...

# Define the target: test-functional
test-functional:
	go test -tags=functional -v ./...

# Define the target: test
# run all the test in current directory and all subdirectories
test:
	go test -tags=unit,functional -v ./...

# Define the target: start-local-infra
start-local-infra:
	docker compose -f ./dev/infra/docker-compose.yml up -d

# Define the target: stop-local-infra
stop-local-infra:
	docker compose -f ./dev/infra/docker-compose.yml down

# Define the target: clean
clean:
	rm -rf bin
	echo "Clean completed"

# Define the target: tidy (clean up go.mod and go.sum)
tidy:
	go mod tidy
