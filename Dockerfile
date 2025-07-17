# --- Build Stage ---
# This stage compiles the Go application.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the Go module files and download dependencies.
# This is done as a separate step to leverage Docker layer caching.
COPY go.mod ./
RUN go mod download

# Copy the rest of the application source code.
COPY . .

# Build the application, creating a static binary.
# CGO_ENABLED=0 is important for creating a static binary that can run in a minimal container.
# -ldflags "-w -s" strips debugging information, reducing the binary size.
RUN CGO_ENABLED=0 go build -ldflags "-w -s" -o /syac

# --- Final Stage ---
# This stage creates the final, lightweight image.
FROM alpine:latest
RUN apk add --no-cache docker-cli git

# Copy the compiled binary from the builder stage.
COPY --from=builder /syac /usr/local/bin/syac

# Set the entrypoint for the container. When the container runs, it will execute the syac binary.

