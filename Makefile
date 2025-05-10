BINARY_NAME=qbit
VERSION=0.0.1
#VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -w -s"
PLATFORMS=darwin-arm64 linux-amd64 windows-amd64

.PHONY: build clean release build-all deps compress

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) main.go

build-all:
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS = $(firstword $(subst -, ,$(platform)))) \
		$(eval ARCH = $(lastword $(subst -, ,$(platform)))) \
		$(eval OUTPUT = bin/$(BINARY_NAME)-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.exe)) \
		echo "Building $(OUTPUT)..."; \
		CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(OUTPUT) main.go; \
		echo "Building $(OUTPUT) done!"; \
    )

compress:
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS = $(firstword $(subst -, ,$(platform)))) \
		$(eval ARCH = $(lastword $(subst -, ,$(platform)))) \
		$(eval BIN = bin/$(BINARY_NAME)-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.exe)) \
		$(eval PKG = bin/$(BINARY_NAME)-$(VERSION)-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.zip,.gz)) \
		echo "Compressing $(BIN) -> $(PKG)"; \
		$(if $(filter windows,$(OS)), \
			zip -FS -j $(PKG) $(BIN) && rm $(BIN), \
			gzip -f -c $(BIN) > $(PKG) && rm $(BIN) \
		); \
		echo "Compressing $(BIN) -> $(PKG) done!"; \
	)
	@cd bin && shasum -a 256 *.gz *.zip > sha256sums.txt && cat sha256sums.txt

clean:
	rm -rf bin/

release: build-all compress

deps:
	go mod tidy
