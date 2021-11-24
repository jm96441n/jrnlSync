.PHONY: .run
run:
	go build -ldflags="-X 'main.Version=0.1.0'" .
	cp ./jrnlNotion ~/bin
