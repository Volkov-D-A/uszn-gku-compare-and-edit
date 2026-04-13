SHELL := /bin/bash

ROOT_DIR := $(abspath .)
FRONTEND_DIR := $(ROOT_DIR)/frontend
PKGCONFIG_DIR := $(ROOT_DIR)/build/pkgconfig
GOCACHE_DIR := /tmp/uszn-gku-go-cache

export GOCACHE := $(GOCACHE_DIR)
export PKG_CONFIG_PATH := $(PKGCONFIG_DIR):$(PKG_CONFIG_PATH)

.PHONY: help frontend-install frontend-build test dev build build-windows cli clean

help:
	@echo "Available targets:"
	@echo "  make frontend-install  - install frontend dependencies"
	@echo "  make frontend-build    - build frontend assets into build/frontend"
	@echo "  make test              - run Go tests"
	@echo "  make dev               - run Wails dev"
	@echo "  make build             - build desktop app with Wails"
	@echo "  make build-windows     - build Windows desktop app with Wails"
	@echo "  make cli               - run CLI analysis against test DBF files"
	@echo "  make clean             - remove generated assets and caches from the repo"

frontend-install:
	npm --prefix "$(FRONTEND_DIR)" install

frontend-build: frontend-install
	npm --prefix "$(FRONTEND_DIR)" run build

test:
	go test ./...

dev: frontend-install
	wails dev

build: frontend-build test
	wails build -s

build-windows: frontend-build test
	wails build -s -platform windows/amd64

cli:
	go run . --cli test_data/chrg_356_92_202601.dbf test_data/chrg_356_92_202602.dbf 20 build/report.xlsx

clean:
	rm -rf "$(ROOT_DIR)/build/frontend/assets"
	rm -f "$(ROOT_DIR)/build/frontend/index.html"
