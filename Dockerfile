FROM ubuntu:22.04 AS builder

# Install dependencies

RUN apt-get update && apt-get install -y \
    build-essential \
    libvips-dev && \
    rm -rf /var/lib/apt/lists/*


    # Install Go
ADD https://go.dev/dl/go1.23.3.linux-amd64.tar.gz /usr/local/
RUN tar -C /usr/local -xzf /usr/local/go1.23.3.linux-amd64.tar.gz && \
    rm /usr/local/go1.23.3.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOBIN="/go/bin"
ENV GOROOT="/usr/local/go"

WORKDIR /app

# Copy the source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application

RUN go build -o main ./cmd/api/main.go

FROM ubuntu:22.04

WORKDIR /app

# Install only necessary runtime dependencies
RUN apt update && apt install -y libvips-dev && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main .
EXPOSE 8080

CMD ["./main"]