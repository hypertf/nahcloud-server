# NahCloud

A fake cloud API for testing Terraform tooling without provisioning real infrastructure.

## Why

You want to test your Terraform providers, CI/CD pipelines, or automation tooling. You don't want to:
- Pay for real cloud resources
- Wait for slow API calls
- Deal with rate limits and quotas
- Clean up orphaned resources

NahCloud gives you a local cloud-shaped API that accepts your requests, stores state in SQLite, and lets you move on with your life.

## How it's different from LocalStack

**LocalStack** mocks AWS. It's great if you're testing AWS-specific Terraform configs.

**NahCloud** is a generic fake cloud with simple, predictable resources. It's for testing:
- Terraform provider development
- CI/CD pipeline logic
- Retry and error handling (via chaos engineering)
- State backend behavior

## Features

### Simple Resources
- **Projects** - top-level containers
- **Instances** - compute resources with CPU, memory, image, status
- **Metadata** - key-value storage with path-based hierarchy
- **Buckets & Objects** - blob storage

### Terraform State Backend
NahCloud implements the Terraform HTTP state backend protocol:
- `GET/POST/DELETE /v1/tfstate/{id}` - state operations
- `LOCK/UNLOCK /v1/tfstate/{id}` - state locking

### Chaos Engineering
Inject failures to test how your tooling handles a misbehaving API:

```bash
# Enable chaos mode
NAH_CHAOS_ENABLED=true

# Add latency (min-max milliseconds)
NAH_LATENCY_GLOBAL_MS=10-100
NAH_LATENCY_PROJECTS_MS=50-200

# Inject errors (0.0 to 1.0)
NAH_ERRRATE_PROJECTS=0.1      # 10% of project calls fail
NAH_ERRRATE_INSTANCES=0.05
NAH_ERRRATE_METADATA=0.05

# Error types and weights
NAH_ERROR_TYPES=503,500,429   # which errors to return
NAH_ERROR_WEIGHTS=3,2,1       # relative frequency
```

Bypass chaos per-request with headers:
- `X-Nah-No-Chaos: true` - skip all chaos
- `X-Nah-Latency: 50` - force specific latency (ms)

### Web Console
Browse and manage resources at `http://localhost:8080/web/`

## Quick Start

```bash
# Run the server
make run-server

# Or with chaos enabled
make run-server-with-chaos
```

The API is available at `http://localhost:8080/v1/`

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NAH_HTTP_ADDR` | `:8080` | Server listen address |
| `NAH_TOKEN` | (none) | Bearer token for auth (optional) |
| `NAH_SQLITE_DSN` | `file:nah.db?...` | SQLite connection string |

## API Overview

```
# Projects
POST   /v1/projects
GET    /v1/projects
GET    /v1/projects/{id}
PATCH  /v1/projects/{id}
DELETE /v1/projects/{id}

# Instances
POST   /v1/instances
GET    /v1/instances
GET    /v1/instances/{id}
PATCH  /v1/instances/{id}
DELETE /v1/instances/{id}

# Metadata
POST   /v1/metadata
GET    /v1/metadata?prefix=...
GET    /v1/metadata/{id}
PATCH  /v1/metadata/{id}
DELETE /v1/metadata/{id}

# Buckets
POST   /v1/buckets
GET    /v1/buckets
GET    /v1/buckets/{id}
PATCH  /v1/buckets/{id}
DELETE /v1/buckets/{id}

# Objects
POST   /v1/bucket/{bucket_id}/objects
GET    /v1/bucket/{bucket_id}/objects?prefix=...
GET    /v1/bucket/{bucket_id}/objects/{id}
PATCH  /v1/bucket/{bucket_id}/objects/{id}
DELETE /v1/bucket/{bucket_id}/objects/{id}

# Terraform State
GET    /v1/tfstate/{id}
POST   /v1/tfstate/{id}
DELETE /v1/tfstate/{id}
LOCK   /v1/tfstate/{id}
UNLOCK /v1/tfstate/{id}
```

## License

MIT
