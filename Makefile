GO_BUILD_FLAGS = -ldflags "-s -w"
CGO_ENABLED = 0

.PHONY: all build clean run

all: build

build:
	@echo "Building codecompass..."
	@CGO_ENABLED=$(CGO_ENABLED) go build $(GO_BUILD_FLAGS) -o codecompass main.go
	@echo "Build complete. Executable: ./codecompass"

clean:
	@echo "Cleaning up..."
	@rm -f codecompass
	@echo "Clean complete."

run:
	@echo "Running codecompass..."
	@./codecompass
