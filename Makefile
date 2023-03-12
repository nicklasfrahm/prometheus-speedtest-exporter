REGISTRY	:= ghcr.io
REPO		:= nicklasfrahm/prometheus-speedtest-exporter
TARGET		?= prometheus-speedtest-exporter
SOURCES		:= $(shell find . -name "*.go")
PLATFORM	?= $(shell go version | cut -d " " -f 4)
GOOS		:= $(shell echo $(PLATFORM) | cut -d "/" -f 1)
GOARCH		:= $(shell echo $(PLATFORM) | cut -d "/" -f 2)
SUFFIX		:= $(GOOS)-$(GOARCH)
VERSION		?= $(shell git describe --always --tags --dirty)
BUILD_FLAGS	:= -ldflags="-s -w -X main.version=$(VERSION)"

ifeq ($(GOOS),windows)
SUFFIX	= $(GOOS)-$(GOARCH).exe
endif

BINARY		?= bin/$(TARGET)-$(SUFFIX)

# Global default target.
.PHONY: all
all: build

# Build the binary.
run: bin/$(TARGET)-$(SUFFIX)
	./bin/$(TARGET)-$(SUFFIX)

build: bin/$(TARGET)-$(SUFFIX)

bin/$(TARGET)-$(SUFFIX): $(SOURCES)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(BINARY) cmd/$(TARGET)/main.go
ifdef UPX
	upx -qq $(UPX) $(BINARY)
endif

# Build container image.
.PHONY: docker
docker:
	docker build \
		-t $(TARGET):latest \
		-t $(TARGET):$(VERSION) \
		-t $(REGISTRY)/$(REPO)-$(TARGET):latest \
		-t $(REGISTRY)/$(REPO)-$(TARGET):$(VERSION) \
		--build-arg TARGET=$(TARGET) \
		--build-arg VERSION=$(VERSION) \
		-f build/package/Dockerfile .
