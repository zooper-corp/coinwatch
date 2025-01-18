FROM golang:1.22.3 AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY . .
RUN go mod download
ARG TARGETARCH
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        apt-get install -y --no-install-recommends gcc-aarch64-linux-gnu libc6-dev-arm64-cross; \
        CC=aarch64-linux-gnu-gcc; \
    elif [ "$TARGETARCH" = "amd64" ]; then \
        apt-get install -y --no-install-recommends gcc-x86-64-linux-gnu libc6-dev-amd64-cross; \
        CC=x86_64-linux-gnu-gcc; \
    fi && \
    CGO_ENABLED=1 CC=$CC GOOS=linux GOARCH=$TARGETARCH \
    go build -o /coinwatch -a -ldflags '-linkmode external -extldflags "-static"' . && \
    apt-get remove -y gcc-aarch64-linux-gnu gcc-x86-64-linux-gnu libc6-dev-arm64-cross libc6-dev-amd64-cross && \
    rm -rf /var/lib/apt/lists/*

# Create a smaller runtime image
FROM debian:buster-slim
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /coinwatch /coinwatch
EXPOSE 8080
ENTRYPOINT ["/coinwatch"]
