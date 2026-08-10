---
name: a6-recipe-multi-tenant
description: >-
  Recipe skill for implementing multi-tenant API gateway patterns using the a6
  CLI. Covers tenant isolation via Consumer Groups, host/path/header-based
  routing, per-tenant rate limiting, context forwarding with proxy-rewrite,
  and declarative config sync workflows for multi-tenant management.
version: "1.0.0"
author: Apache APISIX Contributors
license: Apache-2.0
metadata:
  category: recipe
  apisix_version: ">=3.11.0"
  a6_commands:
    - a6 consumer create
    - a6 consumer-group create
    - a6 credential create
    - a6 route create
    - a6 route update
    - a6 config sync
    - a6 config dump
---

# a6-recipe-multi-tenant

## Overview

Multi-tenancy in an API gateway means serving multiple isolated tenants (customers,
teams, or business units) through the same gateway instance, each with their own
rate limits, authentication, and routing rules.

APISIX achieves multi-tenancy through:
1. **Consumer Groups** — group consumers into tenants with shared plugin configs
2. **Host/path/header-based routing** — route requests to tenant-specific upstreams
3. **Per-tenant rate limiting** — enforce quotas per consumer group
4. **Proxy-rewrite** — forward tenant context to backends via headers

## When to Use

- Multiple customers sharing a single API gateway
- Internal platform serving different teams with isolated quotas
- SaaS application requiring per-tenant rate limits and auth
- Need to forward tenant identity to backend services

## Approach A: Consumer Groups for Tenant Isolation

Group consumers by tenant. Each tenant gets shared plugin configuration
(rate limits, transformations) applied via the consumer group.

### 1. Create consumer groups (one per tenant)

```bash
# Free tier — 100 requests/day
a6 consumer-group create -f - <<'EOF'
{
  "id": "tenant-free",
  "desc": "Free tier tenant",
  "plugins": {
    "limit-count": {
      "count": 100,
      "time_window": 86400,
      "key_type": "var",
      "key": "consumer_name",
      "rejected_code": 429,
      "rejected_msg": "Free tier quota exceeded"
    }
  }
}
EOF

# Pro tier — 10000 requests/day
a6 consumer-group create -f - <<'EOF'
{
  "id": "tenant-pro",
  "desc": "Pro tier tenant",
  "plugins": {
    "limit-count": {
      "count": 10000,
      "time_window": 86400,
      "key_type": "var",
      "key": "consumer_name",
      "rejected_code": 429,
      "rejected_msg": "Pro tier quota exceeded"
    }
  }
}
EOF
```

### 2. Create consumers assigned to groups

```bash
a6 consumer create -f - <<'EOF'
{
  "username": "acme-corp",
  "group_id": "tenant-pro",
  "plugins": {
    "key-auth": { "key": "acme-secret-key" }
  }
}
EOF

a6 consumer create -f - <<'EOF'
{
  "username": "startup-xyz",
  "group_id": "tenant-free",
  "plugins": {
    "key-auth": { "key": "startup-xyz-key" }
  }
}
EOF
```

### 3. Create a shared route with auth

```bash
a6 route create -f - <<'EOF'
{
  "id": "api-v1",
  "uri": "/api/v1/*",
  "upstream": {
    "type": "roundrobin",
    "nodes": { "api-backend:8080": 1 }
  },
  "plugins": {
    "key-auth": {}
  }
}
EOF
```

Now `acme-corp` gets 10,000 req/day and `startup-xyz` gets 100 req/day,
both through the same route.

## Approach B: Host-Based Tenant Routing

Route each tenant to their own backend based on the `Host` header.

### 1. Create per-tenant upstreams

```bash
a6 upstream create -f - <<'EOF'
{
  "id": "upstream-tenant-a",
  "type": "roundrobin",
  "nodes": { "tenant-a-backend:8080": 1 }
}
EOF

a6 upstream create -f - <<'EOF'
{
  "id": "upstream-tenant-b",
  "type": "roundrobin",
  "nodes": { "tenant-b-backend:8080": 1 }
}
EOF
```

### 2. Create host-based routes

```bash
a6 route create -f - <<'EOF'
{
  "id": "tenant-a-route",
  "host": "tenant-a.example.com",
  "uri": "/*",
  "upstream_id": "upstream-tenant-a",
  "plugins": { "key-auth": {} }
}
EOF

a6 route create -f - <<'EOF'
{
  "id": "tenant-b-route",
  "host": "tenant-b.example.com",
  "uri": "/*",
  "upstream_id": "upstream-tenant-b",
  "plugins": { "key-auth": {} }
}
EOF
```

## Approach C: Header-Based Tenant Routing

Use a custom header (e.g., `X-Tenant-ID`) to route to different upstreams
via `traffic-split`.

```bash
a6 route create -f - <<'EOF'
{
  "uri": "/api/*",
  "plugins": {
    "key-auth": {},
    "traffic-split": {
      "rules": [
        {
          "match": [{ "vars": [["http_x_tenant_id", "==", "tenant-a"]] }],
          "weighted_upstreams": [
            { "upstream": { "type": "roundrobin", "nodes": { "tenant-a-backend:8080": 1 } }, "weight": 1 }
          ]
        },
        {
          "match": [{ "vars": [["http_x_tenant_id", "==", "tenant-b"]] }],
          "weighted_upstreams": [
            { "upstream": { "type": "roundrobin", "nodes": { "tenant-b-backend:8080": 1 } }, "weight": 1 }
          ]
        }
      ]
    }
  },
  "upstream": {
    "type": "roundrobin",
    "nodes": { "default-backend:8080": 1 }
  }
}
EOF
```

## Forwarding Tenant Context to Backends

Use `proxy-rewrite` to inject tenant identity as headers so backends
know which tenant the request belongs to.

```bash
a6 route update api-v1 -f - <<'EOF'
{
  "plugins": {
    "key-auth": {},
    "proxy-rewrite": {
      "headers": {
        "set": {
          "X-Consumer-Name": "$consumer_name",
          "X-Consumer-Group": "$consumer_group_id"
        }
      }
    }
  }
}
EOF
```

Backend receives `X-Consumer-Name: acme-corp` and `X-Consumer-Group: tenant-pro`.

## Declarative Multi-Tenant Config

Manage tenant groups, consumers, and routes declaratively with `a6 config sync`:

```yaml
# apisix-tenants.yaml
consumer_groups:
  - id: tenant-free
    desc: "Free tier"
    plugins:
      limit-count:
        count: 100
        time_window: 86400
        key_type: var
        key: consumer_name
  - id: tenant-pro
    desc: "Pro tier"
    plugins:
      limit-count:
        count: 10000
        time_window: 86400
        key_type: var
        key: consumer_name

consumers:
  - username: acme-corp
    group_id: tenant-pro
  - username: startup-xyz
    group_id: tenant-free

routes:
  - id: api-v1
    uri: "/api/v1/*"
    upstream:
      type: roundrobin
      nodes:
        "api-backend:8080": 1
    plugins:
      key-auth: {}
      proxy-rewrite:
        headers:
          set:
            X-Consumer-Name: "$consumer_name"
            X-Consumer-Group: "$consumer_group_id"
```

```bash
# Preview changes
a6 config diff -f apisix-tenants.yaml

# Apply
a6 config sync -f apisix-tenants.yaml
```

Create each tenant's `key-auth` data as a credential after the consumers
exist. For example, save the following as `acme-credential.yaml`:

```yaml
id: acme-key-auth
plugins:
  key-auth:
    key: acme-secret-key
```

```bash
a6 credential create --consumer acme-corp -f acme-credential.yaml
```

Save the free-tier credential as `startup-credential.yaml`:

```yaml
id: startup-key-auth
plugins:
  key-auth:
    key: startup-xyz-key
```

```bash
a6 credential create --consumer startup-xyz -f startup-credential.yaml
```

## Gotchas

- **Credentials are separate resources** — `a6 config sync` and `a6 config dump`
  do not manage Consumer Credential subresources. Store credential files securely
  and apply or restore them separately with `a6 credential` commands.
- **Consumer group plugins merge** — plugins set on the consumer group are merged
  with plugins on the individual consumer. The consumer's plugin config takes
  precedence if both define the same plugin.
- **`group_id` is a string** — must match an existing consumer group ID exactly.
- **Rate limit key** — use `key_type: "var"` with `key: "consumer_name"` to
  enforce per-consumer limits within a group. Without this, the limit applies
  globally across all consumers in the group.
- **Variable names in proxy-rewrite** — `$consumer_name` and `$consumer_group_id`
  are APISIX built-in variables, available only after authentication runs.
  Ensure the auth plugin (key-auth, jwt-auth, etc.) has higher priority than
  proxy-rewrite.

## Verification

```bash
# List consumer groups
a6 consumer-group list

# Verify consumer assignment
a6 consumer get acme-corp --output json | grep group_id

# Test rate limiting for free tier
for i in $(seq 1 101); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "apikey: startup-xyz-key" http://localhost:9080/api/v1/hello
done
# Request 101 should return 429

# Verify tenant headers reach backend
curl -H "apikey: acme-secret-key" http://localhost:9080/api/v1/headers
# Response should show X-Consumer-Name and X-Consumer-Group headers
```
