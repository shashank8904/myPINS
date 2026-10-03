# Stage 1: Build Go binaries
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the api and worker binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/worker

# Stage 2: Runtime
FROM alpine:3.19

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Copy pre-built frontend assets for the API to serve.
# Note: Ensure you run `cd apps/web && npm install && npm run build` 
# on your host machine before running `docker build`.
COPY apps/web/dist/ ./apps/web/dist/

# Set environment variables for runtime
ENV PORT=8080

EXPOSE 8080

# Default command runs the API server. 
# Run the worker separately via: docker run <image> /app/worker
CMD ["/app/api"]
