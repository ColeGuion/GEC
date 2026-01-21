# docker/Dockerfile

############################
# 1) Builder stage
############################
FROM ubuntu:22.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

# Build deps: Go, gcc/g++, make, and runtime libs needed for linking
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    make \
    gcc-12 g++-12 \
    gcc g++ \
    wget \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set GCC 12 as default
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-12 100 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-12 100

# Download and install Go 1.24.2
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

# Build native runtime library (libgec.a)
RUN make -C src/native/gec_runtime

# Build Go server (CGO enabled)
# If your top-level Makefile already sets CGO flags correctly, you can do: RUN make server
#RUN CGO_ENABLED=1 go build -o /app/build/gec-server ./src/cmd/gec-server
RUN CGO_ENABLED=1 \
    CGO_CFLAGS="-Isrc/native/gec_runtime/include -Isrc/native/gec_runtime/third_party/onnxruntime/include" \
    CGO_LDFLAGS="-Lsrc/native/gec_runtime/build -lgec \
                 -Lsrc/native/gec_runtime/third_party/onnxruntime/lib -lonnxruntime \
                 -lstdc++" \
    go build -o /app/gec-server ./src/cmd/gec-server


############################
# 2) Runtime stage
############################
FROM ubuntu:22.04 AS runtime

ENV DEBIAN_FRONTEND=noninteractive

# Runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy server binary
COPY --from=builder /app/gec-server /app/gec-server

# Copy models + webpage assets + data files used by your Go packages
#COPY --from=builder /app/models /app/models
COPY --from=builder /app/webpage /app/webpage
COPY --from=builder /app/src /app/src

# Copy ONNX Runtime shared libraries
COPY --from=builder /app/src/native/gec_runtime/third_party/onnxruntime/lib/ /usr/local/lib/

# Register shared libraries
RUN ldconfig

# Make the service use stable, repo-relative paths
ENV GEC_ROOT=/app
ENV MODEL_ROOT=/app/models
ENV PORT=8089

EXPOSE 8089

ENTRYPOINT ["/app/gec-server"]
