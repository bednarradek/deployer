build-linux:
	GOOS=linux GOARCH=arm go build -o bin/output

build:
	GOOS=linux GOARCH=arm go build -o bin/output

tidy:
	go mod tidy

upgrade-packages:
	go get -u
	go mod tidy
	go mod vendor

vendor:
	go mod vendor

test:
	go test -v ./...

ling-docker:
	docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run --skip-dirs vendor --modules-download-mode vendor

lint-dockerfile:
	docker run --rm --interactive hadolint/hadolint < Dockerfile

format:
	gofmt -s -l -w internal/ cmd/ *.go