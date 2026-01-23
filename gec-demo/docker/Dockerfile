# docker/Dockerfile

############################
# 1) Builder stage
############################
FROM ubuntu:24.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

# Build deps: Go, gcc/g++, make, and runtime libs needed for linking
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    make \
    gcc-12 g++-12 gcc g++ \
    wget \
    pkg-config \
    libhunspell-dev \
    libjson-c-dev \
    libtcmalloc-minimal4 \
    && rm -rf /var/lib/apt/lists/*

# Set GCC 12 as default
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-12 100 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-12 100

# Download and Install Go 1.24.2
RUN wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# Add Go to PATH
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app

# Copy Go modules files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY src ./src
COPY webpage ./webpage
#COPY models ./models


# Builds native runtime library (libgec.a)
RUN make -C src/native/gec_runtime

# Build Go server (CGO enabled)
ENV CGO_ENABLED=1
RUN go build -o /app/gec-server ./src/cmd/gec-server


############################
# 2) Runtime stage
############################
FROM ubuntu:24.04 AS runtime

ENV DEBIAN_FRONTEND=noninteractive

# Runtime dependencies
# - libstdc++6: for C++ runtime
# - libhunspell-1.7-0: hunspell runtime shared lib (Ubuntu 24.04)
# - libjson-c5
# - libtcmalloc-minimal4
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libstdc++6 \
    libhunspell-1.7-0 \
    libjson-c5 \
    libtcmalloc-minimal4 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

## Add Python & huggingface_hub CLI to download the model if need be
#? May need to combine with other dependecies right above
#RUN apt-get update && apt-get install -y --no-install-recommends \
#    python3 python3-pip \
#    && pip3 install --no-cache-dir --upgrade huggingface_hub \
#    && rm -rf /var/lib/apt/lists/*
## Copy entrypoint script
#COPY docker/entrypoint.sh /app/entrypoint.sh
#ENTRYPOINT ["/app/entrypoint.sh"]

# Copy webpage assets
COPY --from=builder /app/webpage /app/webpage

# Copy shared libraries
COPY --from=builder /app/src/native/gec_runtime/third_party/onnxruntime/lib/ /usr/local/lib/
COPY --from=builder /app/src/native/gec_runtime/third_party/sentencepiece/lib/ /usr/local/lib/
COPY --from=builder /app/src/native/gec_runtime/third_party/icu/lib/ /usr/local/lib/

# Copy server binary
COPY --from=builder /app/gec-server /app/gec-server

# Copy source config file
#COPY --from=builder /app/src/native/gec_runtime/config /app/src/native/gec_runtime/config
#COPY --from=builder /app/src /app/src
#COPY --from=builder /app/models /app/models


# Copy model files
#COPY --from=builder /app/models/GecModel /models/GecModel

# Register shared libraries
RUN ldconfig

# Stable Runtime Paths
ENV MODEL_ROOT=/app/models
ENV PORT=8089

EXPOSE 8089

ENTRYPOINT ["/app/gec-server"]
