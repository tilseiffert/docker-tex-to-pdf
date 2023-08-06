

# Start from the official Go image
FROM golang:1.20-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies. 
# Dependencies will be cached if the go.mod and go.sum files are not changed 
RUN go mod download 

# Copy the source from the current directory to the Working Directory inside the container 
COPY . .

# # Run the tests
# RUN go test -v

# Build the Go app
RUN go build -o tex-to-pdfa cmd/tex-to-pdfa/main.go

# Start from the custom base image
FROM tex-to-pdfa-base:12-slim

# Copy the binary from builder
COPY --from=builder /app/tex-to-pdfa /usr/local/bin/

# Set the working directory
WORKDIR /data

# Run the script when the container starts
CMD ["/usr/local/bin/tex-to-pdfa"]
