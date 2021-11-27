.PHONY: .run
run:
	go build -ldflags="-X 'main.Version=0.2.0'" .

.PHONY: .cov
cov:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
