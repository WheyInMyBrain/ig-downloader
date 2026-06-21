# Stage 1: Build the standalone Go binary
FROM golang:1.26.3-alpine AS builder

WORKDIR /app

# Copy the dependency specs first to leverage Docker caching
COPY go.mod ./
RUN go mod download || true

# Copy all the rest of the project repository files
COPY . .

# Compile a statically linked, highly optimized standalone binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ig-downloader ./src

# Stage 2: Final minimal runtime execution layer
FROM alpine:3.19

# Install standard CA root security certificates so HTTPS API queries work inside the container
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the compiled binary over from the builder stage
COPY --from=builder /ig-downloader .

# Create the internal target directory where downloads will collect
RUN mkdir /downloads

# Inform Docker that this path maps to external host storage
VOLUME ["/downloads"]

# Expose the local hosting UI gateway port
EXPOSE 8080

# Execute the application, forcing the Output Directory fallback to our Volume anchor point
ENTRYPOINT ["./ig-downloader", "--serve", "--dir=/downloads"]