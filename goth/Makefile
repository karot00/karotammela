.PHONY: build run dev test test-race css fmt vet tidy lint clean release

BINARY := bin/goth
PKG := ./...
RELEASE := goth-$(shell date +%Y%m%d)-linux-amd64

build: css
	go build -o $(BINARY) ./cmd/server

run: css
	go run ./cmd/server

dev: css
	@echo "Run Next.js (:3000), Go (:8080) and Caddy (:80) separately; this target starts the Go server."
	go run ./cmd/server

css:
	./node_modules/.bin/tailwindcss -i web/input.css -o web/static/app.css --minify

test:
	go test $(PKG)

test-race:
	go test -race $(PKG)

fmt:
	gofmt -w .

vet:
	go vet $(PKG)

tidy:
	go mod tidy

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed; skipping"

clean:
	rm -rf bin

# Versioned release artifact (Phase 12.5 exit gate): a single self-contained
# linux/amd64 binary (templates, static assets, media, and migrations are all
# embedded) plus the systemd units and a sha256 manifest.
release: css
	mkdir -p dist/$(RELEASE)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/$(RELEASE)/goth ./cmd/server
	cp -r deploy dist/$(RELEASE)/deploy
	cd dist && tar -czf $(RELEASE).tar.gz $(RELEASE)
	cd dist && sha256sum $(RELEASE).tar.gz > $(RELEASE).tar.gz.sha256
	rm -rf dist/$(RELEASE)
	@echo "dist/$(RELEASE).tar.gz ready"
