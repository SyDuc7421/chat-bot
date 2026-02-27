.PHONY: all tidy vet fmt test swagger build run clean checks pre-push

# Default target
all: checks build

# Run go mod tidy
tidy:
	go mod tidy

# Run go vet
vet:
	go vet ./...

# Run gofmt (formats code)
fmt:
	gofmt -w .

# Run unit tests
test:
	go test -v ./...

# Generate swagger docs
swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init

# Build the application
build: swagger
	go build -o tmp/main .

# Run the application
run: swagger
	go run main.go

# Clean built files
clean:
	rm -rf tmp/

# Run format, tidy, and vet
checks: fmt tidy vet

# Combine formatting, vet, test, and swagger generation before pushing
pre-push: checks test swagger
	@echo "All tests passed! Ready to push."
