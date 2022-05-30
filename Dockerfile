FROM golang:1.18.2 AS builder

RUN dpkg --add-architecture amd64 \
    && apt update \
    && apt-get install -y --no-install-recommends gcc-x86-64-linux-gnu libc6-dev-amd64-cross

WORKDIR /src
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 CC=x86_64-linux-gnu-gcc GOOS=linux GOARCH=amd64 \
    go build -o /coinwatch -a -ldflags '-linkmode external -extldflags "-static"' .

FROM debian:buster-slim
# Copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy binary
COPY --from=builder /coinwatch /coinwatch
EXPOSE 8080

ENTRYPOINT ["/coinwatch"]