build:
	go build -o vellum ./cmd/vellum

run:
	go run ./cmd/vellum

clean:
	rm -f vellum
