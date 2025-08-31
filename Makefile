# Define the output for project's binary
OUTPUT = bin/fury-flow

# Default target
all: build

# Define the target: build
build:
	mkdir -p bin
	go build -o $(OUTPUT) .
	echo "Build completed: $(OUTPUT)"

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
