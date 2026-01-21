# Build arguments
ARG GO_VERSION=1.24
ARG NODE_VERSION=22
ARG VERSION=docker

# Stage 1: Build Go binaries
FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION

WORKDIR /build

# Copy Go module files
COPY src/go.mod src/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./

# Build claudex and claudex-hooks with static linking
# Let Go use the container's native architecture
RUN CGO_ENABLED=0 go build \
    -ldflags "-X main.Version=${VERSION}" \
    -o claudex ./cmd/claudex && \
    CGO_ENABLED=0 go build \
    -o claudex-hooks ./cmd/claudex-hooks

# Stage 2: Runtime image
FROM node:${NODE_VERSION}-alpine

WORKDIR /workspace

# Install Claude CLI and git (for auto-docs feature)
RUN npm install -g @anthropic-ai/claude-code && \
    apk add --no-cache git

# Copy binaries from builder stage
COPY --from=builder /build/claudex /usr/local/bin/claudex
COPY --from=builder /build/claudex-hooks /usr/local/bin/claudex-hooks

# Use existing node user (UID 1000) - standard in node:alpine images
# Create claudex config directories
RUN mkdir -p /home/node/.config/claudex/profiles \
             /home/node/.config/claudex/hooks \
             /workspace/.claudex/sessions \
             /workspace/.claudex/logs

# Copy profiles and hook scripts
COPY src/profiles /home/node/.config/claudex/profiles
COPY src/scripts/proxies/*.sh /home/node/.config/claudex/hooks/
RUN chmod +x /home/node/.config/claudex/hooks/*.sh

# Set ownership
RUN chown -R node:node /home/node /workspace

# Switch to non-root user
USER node

# Volumes for persistence
VOLUME ["/workspace/.claudex/sessions", "/workspace"]

# Set working directory
WORKDIR /workspace

# Use ENTRYPOINT to allow passing arguments
ENTRYPOINT ["claudex"]
