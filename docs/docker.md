# Running Claudex with Docker

This document describes how to run claudex in a Docker container for isolated development environments.

## Quick Start

```bash
# Build the Docker image
make docker-build

# Run claudex in Docker
make docker-run
```

## Prerequisites

### Required Software

- Docker 24.0+ or Docker Desktop
- Docker Compose 2.0+ (included with Docker Desktop)

### Claude CLI Authentication

Claudex requires Claude CLI authentication to function. You must authenticate the Claude CLI on your host system before running claudex in Docker.

```bash
# Install Claude CLI (if not already installed)
npm install -g @anthropic-ai/claude-code

# Authenticate
claude-code auth login

# Verify authentication
ls ~/.config/claude
```

The Docker container will mount your host's `~/.config/claude` directory to access your authentication credentials.

## Building the Image

### Using Make

```bash
make docker-build
```

This builds two tagged images:
- `claudex:latest`
- `claudex:<VERSION>` (from VERSION variable or git tag)

### Using Docker directly

```bash
# Build with default version
docker build -t claudex:latest .

# Build with specific version
docker build -t claudex:1.0.0 --build-arg VERSION=1.0.0 .

# Build with custom version variable
VERSION=1.0.0 docker build -t claudex:1.0.0 .
```

### Multi-Stage Build Details

The Dockerfile uses a multi-stage build:

1. **Build stage** (`golang:1.24-alpine`): Compiles Go binaries with static linking
2. **Runtime stage** (`node:22-alpine`): Installs Claude CLI and includes only runtime dependencies

Final image size: Approximately 200MB

## Running Claudex

### Using Make

The simplest way to run claudex in Docker:

```bash
make docker-run
```

This command:
- Runs interactively with TTY allocation (`-it`)
- Removes the container after exit (`--rm`)
- Mounts the current directory as `/workspace`
- Persists sessions via volume mount
- Provides read-only access to Claude CLI authentication

### Using Docker directly

```bash
docker run -it --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions \
  -v ~/.config/claude:/home/node/.config/claude:ro \
  claudex:latest
```

Flags explained:
- `-it`: Interactive mode with TTY (required for TUI)
- `--rm`: Remove container after exit
- `-v $(pwd):/workspace`: Mount current directory as working directory
- `-v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions`: Persist sessions
- `-v ~/.config/claude:/home/node/.config/claude:ro`: Claude CLI authentication (read-only)

### Using Docker Compose

Docker Compose simplifies container orchestration with pre-configured volume mounts and environment variables.

```bash
# Run interactively
docker compose run --rm claudex

# Run with custom arguments
docker compose run --rm claudex --help
docker compose run --rm claudex start --profile custom
```

The `docker-compose.yml` file is pre-configured with recommended settings.

## Configuration

### Volume Mounts

| Mount | Container Path | Purpose | Required |
|-------|---------------|---------|----------|
| Project files | `/workspace` | Working directory for claudex | Yes |
| Sessions | `/workspace/.claudex/sessions` | Persist session data across container restarts | Yes |
| Claude auth | `/home/node/.config/claude` | Claude CLI authentication credentials | Yes |
| Config file | `/workspace/.claudex/config.toml` | Optional claudex configuration | No |

#### Critical: Session Persistence

Without the sessions volume mount, all session data will be lost when the container stops. Always mount `.claudex/sessions`:

```bash
-v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions
```

### Environment Variables

Claudex behavior can be customized via environment variables:

| Variable | Description | Default | Values |
|----------|-------------|---------|--------|
| `CLAUDEX_AUTODOC_SESSION_PROGRESS` | Enable/disable session progress documentation | `true` | `true`, `false` |
| `CLAUDEX_AUTODOC_SESSION_END` | Enable/disable session end documentation | `true` | `true`, `false` |
| `CLAUDEX_AUTODOC_FREQUENCY` | Number of tool uses between auto-docs | `5` | Integer (1-50) |
| `CLAUDEX_SKIP_DOCS` | Disable all auto-documentation | `false` | `true`, `false` |

#### Setting Environment Variables

With Docker:
```bash
docker run -it --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions \
  -v ~/.config/claude:/home/node/.config/claude:ro \
  -e CLAUDEX_AUTODOC_FREQUENCY=10 \
  -e CLAUDEX_SKIP_DOCS=true \
  claudex:latest
```

With Docker Compose:
```bash
# Set in shell environment
export CLAUDEX_AUTODOC_FREQUENCY=10
docker compose run --rm claudex

# Or inline
CLAUDEX_AUTODOC_FREQUENCY=10 docker compose run --rm claudex
```

## Advanced Usage

### Custom Override File

Docker Compose supports override files for environment-specific customization without modifying the main `docker-compose.yml`.

```bash
# Copy example override file
cp docker-compose.override.yml.example docker-compose.override.yml

# Edit as needed
nano docker-compose.override.yml

# Run with override applied automatically
docker compose run --rm claudex
```

Common customizations in override file:
- Additional volume mounts (SSH keys, custom configs)
- Custom environment variables
- Resource limits (CPU, memory)
- User ID mapping (for permission issues)

### Building with Custom Version

```bash
# Build with specific version
VERSION=1.0.0 make docker-build

# Verify version
docker run --rm claudex:1.0.0 --version
```

### Running with Different User ID

If you encounter permission issues with mounted volumes (common on Linux), run the container with your host user ID:

```bash
# Find your UID
id -u  # Example output: 1001

# Run with custom UID
docker run -it --rm \
  --user 1001:1001 \
  -v $(pwd):/workspace \
  -v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions \
  -v ~/.config/claude:/home/node/.config/claude:ro \
  claudex:latest
```

Or add to `docker-compose.override.yml`:
```yaml
services:
  claudex:
    user: "1001:1001"
```

### Mounting SSH Keys for Git Operations

If your project requires SSH authentication for git operations:

```bash
docker run -it --rm \
  -v $(pwd):/workspace \
  -v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions \
  -v ~/.config/claude:/home/node/.config/claude:ro \
  -v ~/.ssh:/home/node/.ssh:ro \
  claudex:latest
```

Or in `docker-compose.override.yml`:
```yaml
services:
  claudex:
    volumes:
      - ~/.ssh:/home/node/.ssh:ro
```

## Troubleshooting

### TUI not displaying or garbled output

**Symptoms**: Terminal interface doesn't render correctly or shows escape sequences.

**Solution**: Ensure you're using the `-it` flags (interactive + TTY):

```bash
docker run -it --rm claudex:latest
```

Docker Compose includes these by default via `stdin_open: true` and `tty: true`.

### Sessions lost after container restart

**Symptoms**: Previous sessions are not available after restarting the container.

**Solution**: Verify the sessions volume mount is configured:

```bash
-v $(pwd)/.claudex/sessions:/workspace/.claudex/sessions
```

Check the mount exists:
```bash
ls -la .claudex/sessions
```

### Authentication errors

**Symptoms**: "Claude CLI authentication required" or similar errors.

**Solution**:

1. Verify Claude CLI is authenticated on host:
   ```bash
   ls ~/.config/claude
   ```

2. Ensure authentication volume mount is present:
   ```bash
   -v ~/.config/claude:/home/node/.config/claude:ro
   ```

3. Check mount permissions:
   ```bash
   docker run --rm claudex:latest ls -la /home/node/.config/claude
   ```

### Permission denied on mounted volumes

**Symptoms**: Cannot write to `.claudex/sessions` or other mounted directories.

**Solution**:

1. Check ownership of host directory:
   ```bash
   ls -la .claudex/
   ```

2. Run container with your user ID:
   ```bash
   docker run -it --rm --user $(id -u):$(id -g) ...
   ```

3. Or fix ownership of host directory:
   ```bash
   sudo chown -R $(id -u):$(id -g) .claudex/
   ```

### Build context too large or slow builds

**Symptoms**: Docker build uploads many files or takes a long time.

**Solution**:

1. Ensure `.dockerignore` exists and is comprehensive
2. Check what's being included in build context:
   ```bash
   docker build --no-cache --progress=plain -t claudex:latest . 2>&1 | grep "Sending build context"
   ```

3. Add additional patterns to `.dockerignore`:
   ```text
   .git
   .claudex/sessions
   node_modules
   .cache
   dist
   ```

### Container exits immediately

**Symptoms**: Container starts but exits immediately without error.

**Solution**:

1. Check if you're missing the `-it` flags (required for interactive TUI)
2. Try running with explicit command:
   ```bash
   docker run -it --rm claudex:latest --help
   ```

3. Check container logs:
   ```bash
   docker logs <container-id>
   ```

### Image build fails

**Symptoms**: `docker build` fails during Go compilation or npm install.

**Solution**:

1. Ensure you have sufficient disk space
2. Clear Docker build cache:
   ```bash
   docker builder prune
   ```

3. Rebuild from scratch:
   ```bash
   docker build --no-cache -t claudex:latest .
   ```

4. Check Go module dependencies:
   ```bash
   cd src && go mod verify
   ```
