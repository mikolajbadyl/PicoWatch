.PHONY: build build-ui dev-ui dev-go run clean

build: build-ui
	go build -o picowatch .

build-ui:
	cd ui && bun install && bun run build

dev-ui:
	cd ui && bun run dev

dev-go:
	go run .

run: build
	./picowatch

clean:
	rm -f picowatch
	rm -f picowatch.db
	rm -rf ui/build
	mkdir -p ui/build
	@echo '<!DOCTYPE html><html><body>Run make build-ui</body></html>' > ui/build/index.html
