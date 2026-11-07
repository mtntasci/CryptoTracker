APP_NAME := cryptotracker
PKG_NAME := CryptoTracker-Installer.pkg
IDENTIFIER := com.cryptotracker.cli
VERSION := 1.0.0

.PHONY: all build install pkg uninstall test clean

all: build

build:
	go build -ldflags="-s -w" -o $(APP_NAME) ./cmd/crypto-tracker

install:
	@chmod +x install.sh
	@./install.sh

pkg: build
	@STAGE_DIR=$$(mktemp -d); \
	mkdir -p "$$STAGE_DIR/usr/local/bin"; \
	cp $(APP_NAME) "$$STAGE_DIR/usr/local/bin/$(APP_NAME)"; \
	chmod 755 "$$STAGE_DIR/usr/local/bin/$(APP_NAME)"; \
	pkgbuild --root "$$STAGE_DIR" --identifier "$(IDENTIFIER)" --version "$(VERSION)" --install-location "/" "$(PKG_NAME)"; \
	rm -rf "$$STAGE_DIR"

uninstall:
	rm -f /opt/homebrew/bin/$(APP_NAME) /usr/local/bin/$(APP_NAME) $(HOME)/.local/bin/$(APP_NAME)

test:
	go test -v ./...

clean:
	rm -f $(APP_NAME) crypto-tracker $(PKG_NAME)

