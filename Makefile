.PHONY: .run
run:
	go build -ldflags="-X 'main.Version=0.2.0'" .
