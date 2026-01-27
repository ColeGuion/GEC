# gec-demo/Makefile

# ---------- Config ----------
APP_NAME      := gec-server
PORT          ?= 8089

# Paths
NATIVE_DIR    := src/native/gec_runtime
NATIVE_LIB    := $(NATIVE_DIR)/build/libgec.a

GO_CMD_DIR    := src/cmd/gec-server
GO_BIN_DIR    := build
GO_BIN        := $(GO_BIN_DIR)/$(APP_NAME)

# Go environment
GO            := go
GOFLAGS       :=
CGO_ENABLED   := 1

# Export CGO vars so Go can find libgec.a
export CGO_ENABLED := 1
export CGO_CFLAGS  := -I$(NATIVE_DIR)/include
export CGO_LDFLAGS := -L$(NATIVE_DIR)/build -lgec -lstdc++ 

# ---------- Targets ----------
.PHONY: all native server run clean info

all: native server
	@echo "✅ Build complete"

# ---------- Native Runtime ----------
native:
	@echo "🔧 Building native GEC runtime (libgec.a)"
	$(MAKE) -C $(NATIVE_DIR)

# ---------- Go Server ----------
$(GO_BIN): $(NATIVE_LIB)
	@echo "🔨 Building Go server"
	@mkdir -p $(GO_BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o $(GO_BIN) ./$(GO_CMD_DIR)
	@chmod +x $(GO_BIN)
	@file $(GO_BIN) | grep -q "executable" || (echo "ERROR: $(GO_BIN) is not an executable" && file $(GO_BIN) && exit 1)

server: $(GO_BIN)

# ---------- Docker ----------
dock:
	docker compose -f docker-compose.yml up --build

dockBuild:
	docker build -f docker/Dockerfile -t gec-demo:latest .

# Mount the model to the container and run it
dockRun:
	docker run --rm -p 8089:8089 -v "$(shell pwd)/models:/models" gec-demo:latest

dockInteractive: 
	docker run --rm -v "$(shell pwd)/models:/models" -it --entrypoint bash gec-demo:latest

# ---------- Run ----------
run: server
	@echo "🚀 Running $(APP_NAME) on port $(PORT)"
	PORT=$(PORT) ./$(GO_BIN)

# ---------- Info ----------
info:
	@echo "App:        $(APP_NAME)"
	@echo "Port:       $(PORT)"
	@echo "Go binary:  $(GO_BIN)"
	@echo "Native lib: $(NATIVE_LIB)"

# ---------- Clean ----------
clean:
	@echo "🧹 Cleaning native runtime"
	-$(MAKE) -C $(NATIVE_DIR) clean
	@echo "🧹 Cleaning Go build"
	@rm -rf $(GO_BIN_DIR)
