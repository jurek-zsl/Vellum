build:
	go build -o vellum ./cmd/vellum

test:
	go test -v -race ./...

run:
	go run ./cmd/vellum

vet:
	go vet ./...

clean:
	rm -f vellum

