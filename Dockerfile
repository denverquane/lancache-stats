FROM golang:1.17-alpine AS builder

# Git is required for getting the dependencies.
RUN apk add --no-cache git

WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app .

FROM alpine:3.13.0 AS final

WORKDIR /app
# Set up non-root user and app directory
# * Non-root because of the principle of least privlege6
# * App directory to allow mounting volumes
RUN addgroup -g 1000 lancache-stats && \
    adduser -HD -u 1000 -G lancache-stats lancache-stats && \
    chown -R lancache-stats:lancache-stats /app
USER lancache-stats

# Import the compiled executable
COPY --from=builder /app /app

EXPOSE 5000

# Run the compiled binary.
ENTRYPOINT ["./app"]