BINARY_NAME := zswapctl
VERSION := 1.0.0
LDFLAGS := -ldflags "-s -w -X github.com/zswap-go/zswapctl/internal/version.Version=$(VERSION)"

.PHONY: all build clean install uninstall

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/zswapctl
	chmod 755 $(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

install: build
	install -Dm755 $(BINARY_NAME) /usr/bin/$(BINARY_NAME)
	install -Dm644 assets/zswapctl.conf /etc/zswapctl/zswapctl.conf
	install -Dm644 assets/zswapctl.service /etc/systemd/system/zswapctl.service

uninstall:
	rm -f /usr/bin/$(BINARY_NAME)
	rm -f /etc/zswapctl/zswapctl.conf
	rm -f /etc/systemd/system/zswapctl.service
