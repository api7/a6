---
name: a6-plugin-limit-count
description: >-
  Skill for configuring the Apache APISIX limit-count plugin via the a6 CLI.
  Covers fixed-window rate limiting, count/time_window configuration, key types,
  Redis and Redis-cluster policies for distributed limiting, group-based shared
  quotas, consumer-level vs route-level limiting, response headers, and common
  operational patterns.
version: "1.0.0"
author: Apache APISIX Contributors
license: Apache-2.0
metadata:
  category: plugin
  apisix_version: ">=3.0.0"
  plugin_name: limit-count
  a6_commands:
    - a6 route create
    - a6 route update
    - a6 consumer create
    - a6 consumer update
---

# a6-plugin-limit-count

## Overview

The `limit-count` plugin rate-limits requests using a fixed-window counter
algorithm. Define a maximum number of requests (`count`) within a time interval
(`time_window`). Supports per-IP, per-consumer, per-header, or custom variable
keys. For distributed APISIX deployments, use Redis or Redis-cluster as the
shared counter backend.

## When to Use

- Simple request counting (e.g., 100 requests per hour)
- API quota enforcement per consumer or API key
- Shared rate limits across multiple APISIX nodes (via Redis)
- Grouped quotas across multiple routes

## Plugin Configuration Reference

### Core Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `count` | integer | Yes* | — | Max requests allowed in the time window. > 0 |
| `time_window` | integer | Yes* | — | Time window in seconds. > 0 |
| `key_type` | string | No | `"var"` | Key type: `"var"`, `"var_combination"`, or `"constant"` |
| `key` | string | No | `"remote_addr"` | Variable name or combination for counting |
| `rejected_code` | integer | No | `503` | HTTP status on rejection (200–599) |
| `rejected_msg` | string | No | — | Custom rejection message body |
| `group` | string | No | — | Share counters across routes with same group ID |
| `policy` | string | No | `"local"` | Storage: `"local"`, `"redis"`, or `"redis-cluster"` |
| `show_limit_quota_header` | boolean | No | `true` | Include X-RateLimit-* headers in responses |
| `allow_degradation` | boolean | No | `false` | Allow requests when plugin fails |

*Required unless using `rules` array.

### Redis Fields (when `policy: "redis"`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `redis_host` | string | **Yes** | — | Redis server address |
| `redis_port` | integer | No | `6379` | Redis port |
| `redis_username` | string | No | — | Redis ACL username |
| `redis_password` | string | No | — | Redis password |
| `redis_database` | integer | No | `0` | Redis database index |
| `redis_timeout` | integer | No | `1000` | Timeout in milliseconds |
| `redis_ssl` | boolean | No | `false` | Enable TLS to Redis |

### Redis Cluster Fields (when `policy: "redis-cluster"`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `redis_cluster_nodes` | array[string] | **Yes** | — | Array of `"host:port"` (min 2) |
| `redis_cluster_name` | string | **Yes** | — | Cluster name |
| `redis_password` | string | No | — | Cluster password |
| `redis_timeout` | integer | No | `1000` | Timeout in milliseconds |
| `redis_cluster_ssl` | boolean | No | `false` | Enable TLS |

## Key Types

| `key_type` | `key` Format | Example | Description |
|------------|-------------|---------|-------------|
| `"var"` | NGINX variable (no `$`) | `"remote_addr"` | Single variable |
| `"var_combination"` | `$var1 $var2` | `"$remote_addr $consumer_name"` | Multiple variables combined |
| `"constant"` | Any string | `"global"` | Same counter for all requests |

## Response Headers

When `show_limit_quota_header: true` (default):

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Total quota for the time window |
| `X-RateLimit-Remaining` | Remaining requests in current window |
| `X-RateLimit-Reset` | Seconds until counter resets |

## Step-by-Step: Basic Rate Limiting

### 1. Rate limit by client IP (route-level)

```bash
a6 route create -f - <<'EOF'
{
  "id": "rate-limited-api",
  "uri": "/api/*",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 60,
      "key_type": "var",
      "key": "remote_addr",
      "rejected_code": 429,
      "rejected_msg": "Rate limit exceeded. Try again later."
    }
  },
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "backend:8080": 1
    }
  }
}
EOF
```

100 requests per 60 seconds per client IP.

### 2. Rate limit per consumer

```bash
a6 consumer create -f - <<'EOF'
{
  "username": "free-tier",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 3600,
      "rejected_code": 429
    }
  }
}
EOF

a6 consumer create -f - <<'EOF'
{
  "username": "premium",
  "plugins": {
    "limit-count": {
      "count": 10000,
      "time_window": 3600,
      "rejected_code": 429
    }
  }
}
EOF
```

Consumer-level limits apply across all routes the consumer accesses.

## Common Patterns

### Shared quota across routes (group)

```json
{
  "plugins": {
    "limit-count": {
      "count": 1000,
      "time_window": 3600,
      "group": "api-v1",
      "rejected_code": 429
    }
  }
}
```

All routes with `"group": "api-v1"` share the same 1000 req/hour counter.
**Important**: All routes in a group must have identical `limit-count` config.

### Multi-variable key (IP + consumer)

```json
{
  "plugins": {
    "limit-count": {
      "count": 50,
      "time_window": 60,
      "key_type": "var_combination",
      "key": "$remote_addr $consumer_name",
      "rejected_code": 429
    }
  }
}
```

### Global rate limit (all requests share one counter)

```json
{
  "plugins": {
    "limit-count": {
      "count": 10000,
      "time_window": 60,
      "key_type": "constant",
      "key": "global",
      "rejected_code": 429
    }
  }
}
```

### Distributed rate limiting with Redis

```json
{
  "plugins": {
    "limit-count": {
      "count": 1000,
      "time_window": 60,
      "key": "remote_addr",
      "policy": "redis",
      "redis_host": "redis.example.com",
      "redis_port": 6379,
      "redis_password": "secret",
      "redis_database": 0,
      "redis_ssl": true,
      "rejected_code": 429
    }
  }
}
```

Use Redis when running multiple APISIX nodes to share counters.

### Redis cluster

```json
{
  "plugins": {
    "limit-count": {
      "count": 1000,
      "time_window": 60,
      "key": "remote_addr",
      "policy": "redis-cluster",
      "redis_cluster_nodes": [
        "192.168.1.10:6379",
        "192.168.1.11:6379",
        "192.168.1.12:6379"
      ],
      "redis_cluster_name": "apisix-cluster",
      "redis_password": "secret",
      "rejected_code": 429
    }
  }
}
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Limits not shared across APISIX nodes | Using `policy: "local"` (default) | Switch to `"redis"` or `"redis-cluster"` |
| Group config rejected | Mismatched configs in same group | Ensure all routes in group have identical limit-count config |
| Unexpected counter reset | Fixed-window boundary | Normal behavior — counters reset at fixed intervals |
| Key empty, all clients share one counter | Variable doesn't exist | Verify key variable name; falls back to `remote_addr` |
| Rate limit headers missing | `show_limit_quota_header: false` | Set to `true` (default) |
| 503 instead of 429 | Default `rejected_code` is 503 | Set `rejected_code: 429` explicitly |

## Fixed-Window Algorithm Note

`limit-count` uses a fixed-window algorithm. Counters reset at exact intervals.
This means a burst at the boundary of two windows can temporarily exceed the
intended rate (e.g., 100 req/min allows 200 requests if 100 come at t=59s and
100 at t=61s). For smoother rate limiting, combine with `limit-req` (leaky
bucket).

## Config Sync Example

```yaml
version: "1"
consumers:
  - username: free-tier
    plugins:
      limit-count:
        count: 100
        time_window: 3600
        rejected_code: 429
routes:
  - id: rate-limited-api
    uri: /api/*
    plugins:
      limit-count:
        count: 1000
        time_window: 60
        key: remote_addr
        rejected_code: 429
    upstream_id: api-upstream
upstreams:
  - id: api-upstream
    type: roundrobin
    nodes:
      "backend:8080": 1
```
