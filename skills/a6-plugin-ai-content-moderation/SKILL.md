---
name: a6-plugin-ai-content-moderation
description: >-
  Skill for configuring APISIX AI content moderation plugins via the a6 CLI.
  Covers both ai-aws-content-moderation (AWS Comprehend, request-only) and
  ai-aliyun-content-moderation (Aliyun, request + response with streaming),
  toxicity thresholds, category filtering, and integration with ai-proxy.
version: "1.0.0"
author: Apache APISIX Contributors
license: Apache-2.0
metadata:
  category: plugin
  apisix_version: ">=3.9.0"
  plugin_name: ai-aws-content-moderation
  related_plugins:
    - ai-aliyun-content-moderation
  a6_commands:
    - a6 route create
    - a6 route update
    - a6 config sync
---

# a6-plugin-ai-content-moderation

## Overview

APISIX provides two content moderation plugins that filter harmful content
in LLM requests and responses:

| Plugin | Provider | Request | Response | Streaming |
|--------|----------|---------|----------|-----------|
| `ai-aws-content-moderation` | AWS Comprehend | ✅ | ❌ | ❌ |
| `ai-aliyun-content-moderation` | Aliyun Moderation Plus | ✅ | ✅ | ✅ |

Both must be used alongside `ai-proxy` or `ai-proxy-multi`.

## When to Use

- Block toxic, hateful, or sexual content before it reaches the LLM
- Filter harmful LLM responses before they reach clients (Aliyun only)
- Enforce content policies with configurable thresholds
- Comply with content safety regulations

## Plugin Execution Order

```
ai-prompt-template           (priority 1071)
ai-prompt-decorator          (priority 1070)
ai-aws-content-moderation    (priority 1050) ← runs BEFORE ai-proxy
ai-proxy                     (priority 1040)
ai-aliyun-content-moderation (priority 1029) ← runs AFTER ai-proxy
```

The AWS plugin blocks requests before they reach the LLM. The Aliyun plugin
runs after `ai-proxy` sets context and can check both requests and responses.

---

## Plugin 1: ai-aws-content-moderation

Uses the AWS Comprehend `detectToxicContent` API to score request content.

### Configuration Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `comprehend.access_key_id` | string | **Yes** | — | AWS access key ID |
| `comprehend.secret_access_key` | string | **Yes** | — | AWS secret access key |
| `comprehend.region` | string | **Yes** | — | AWS region (e.g. `us-east-1`) |
| `comprehend.endpoint` | string | No | Auto | Custom Comprehend endpoint |
| `comprehend.ssl_verify` | boolean | No | `true` | Verify SSL certificate |
| `moderation_categories` | object | No | — | Per-category thresholds (0-1) |
| `moderation_threshold` | number | No | `0.5` | Overall toxicity threshold (0-1) |

### Moderation Categories

| Category | Description |
|----------|-------------|
| `PROFANITY` | Profane language |
| `HATE_SPEECH` | Hateful content |
| `INSULT` | Insulting language |
| `HARASSMENT_OR_ABUSE` | Harassment or abusive content |
| `SEXUAL` | Sexual content |
| `VIOLENCE_OR_THREAT` | Violent or threatening content |

Each category accepts a score threshold from `0` (strictest, blocks nearly
everything) to `1` (most permissive). If `moderation_categories` is set,
each category is checked individually. Otherwise, the `moderation_threshold`
is used as an overall toxicity check.

### Step-by-Step: AWS Content Moderation

```bash
a6 route create -f - <<'EOF'
{
  "id": "moderated-chat",
  "uri": "/v1/chat/completions",
  "methods": ["POST"],
  "plugins": {
    "ai-aws-content-moderation": {
      "comprehend": {
        "access_key_id": "AKIAIOSFODNN7EXAMPLE",
        "secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
        "region": "us-east-1"
      },
      "moderation_categories": {
        "HATE_SPEECH": 0.3,
        "VIOLENCE_OR_THREAT": 0.2,
        "SEXUAL": 0.5
      }
    },
    "ai-proxy": {
      "provider": "openai",
      "auth": {
        "header": {
          "Authorization": "Bearer sk-your-key"
        }
      },
      "options": {
        "model": "gpt-4"
      }
    }
  }
}
EOF
```

Toxic requests are rejected with HTTP 400:

```
request body exceeds HATE_SPEECH threshold
```

### Overall threshold (no per-category filtering)

```json
{
  "plugins": {
    "ai-aws-content-moderation": {
      "comprehend": {
        "access_key_id": "AKIA...",
        "secret_access_key": "secret...",
        "region": "us-east-1"
      },
      "moderation_threshold": 0.7
    }
  }
}
```

---

## Plugin 2: ai-aliyun-content-moderation

Uses Aliyun Machine-Assisted Moderation Plus. Supports request moderation,
response moderation, and real-time streaming moderation.

### Configuration Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `endpoint` | string | **Yes** | — | Aliyun service endpoint URL |
| `region_id` | string | **Yes** | — | Aliyun region (e.g. `cn-shanghai`) |
| `access_key_id` | string | **Yes** | — | Aliyun access key ID |
| `access_key_secret` | string | **Yes** | — | Aliyun access key secret |
| `check_request` | boolean | No | `true` | Enable request moderation |
| `check_response` | boolean | No | `false` | Enable response moderation |
| `stream_check_mode` | string | No | `final_packet` | `realtime` or `final_packet` |
| `stream_check_cache_size` | integer | No | `128` | Max chars per batch (realtime) |
| `stream_check_interval` | number | No | `3` | Seconds between batch checks (realtime) |
| `request_check_service` | string | No | `llm_query_moderation` | Aliyun service for request checks |
| `request_check_length_limit` | number | No | `2000` | Max chars per request chunk |
| `response_check_service` | string | No | `llm_response_moderation` | Aliyun service for response checks |
| `response_check_length_limit` | number | No | `5000` | Max chars per response chunk |
| `risk_level_bar` | string | No | `high` | Threshold: `none`, `low`, `medium`, `high`, `max` |
| `deny_code` | number | No | `200` | HTTP status code for rejected content |
| `deny_message` | string | No | — | Custom rejection message |
| `timeout` | integer | No | `10000` | Request timeout (ms) |
| `ssl_verify` | boolean | No | `true` | Verify SSL certificate |

### Risk Level System

Content is blocked when its risk level meets or exceeds the `risk_level_bar`:

```
none (0) < low (1) < medium (2) < high (3) < max (4)
```

Setting `risk_level_bar: "high"` blocks content rated `high` or `max`.
Setting `risk_level_bar: "low"` blocks everything rated `low` or above.

### Streaming Modes

| Mode | Behavior |
|------|----------|
| `final_packet` | Buffers entire response, checks at end |
| `realtime` | Checks content in batches during streaming, can interrupt mid-response |

### Step-by-Step: Aliyun Request + Response Moderation

```bash
a6 route create -f - <<'EOF'
{
  "id": "aliyun-moderated-chat",
  "uri": "/v1/chat/completions",
  "methods": ["POST"],
  "plugins": {
    "ai-proxy": {
      "provider": "openai",
      "auth": {
        "header": {
          "Authorization": "Bearer sk-your-key"
        }
      },
      "options": {
        "model": "gpt-4"
      }
    },
    "ai-aliyun-content-moderation": {
      "endpoint": "https://green.cn-shanghai.aliyuncs.com",
      "region_id": "cn-shanghai",
      "access_key_id": "your-aliyun-key-id",
      "access_key_secret": "your-aliyun-key-secret",
      "check_request": true,
      "check_response": true,
      "risk_level_bar": "high",
      "deny_code": 400,
      "deny_message": "Content policy violation"
    }
  }
}
EOF
```

### Realtime streaming moderation

```json
{
  "plugins": {
    "ai-aliyun-content-moderation": {
      "endpoint": "https://green.cn-shanghai.aliyuncs.com",
      "region_id": "cn-shanghai",
      "access_key_id": "key-id",
      "access_key_secret": "key-secret",
      "check_request": true,
      "check_response": true,
      "stream_check_mode": "realtime",
      "stream_check_cache_size": 256,
      "stream_check_interval": 2,
      "risk_level_bar": "medium"
    }
  }
}
```

---

## Integration Patterns

### Pattern A: Request-only filtering (AWS)

```
Client → [AWS Comprehend blocks toxic] → ai-proxy → LLM → Response → Client
```

```yaml
plugins:
  ai-aws-content-moderation:
    comprehend:
      access_key_id: "${AWS_ACCESS_KEY_ID}"
      secret_access_key: "${AWS_SECRET_ACCESS_KEY}"
      region: us-east-1
    moderation_threshold: 0.5
  ai-proxy:
    provider: openai
    auth:
      header:
        Authorization: "Bearer ${OPENAI_API_KEY}"
```

### Pattern B: Request + response filtering (Aliyun)

```
Client → ai-proxy [sets context] → [Aliyun checks request] → LLM
       → [Aliyun checks response] → Client
```

```yaml
plugins:
  ai-proxy:
    provider: openai
    auth:
      header:
        Authorization: "Bearer ${OPENAI_API_KEY}"
  ai-aliyun-content-moderation:
    endpoint: "https://green.cn-shanghai.aliyuncs.com"
    region_id: cn-shanghai
    access_key_id: "${ALIYUN_KEY_ID}"
    access_key_secret: "${ALIYUN_KEY_SECRET}"
    check_request: true
    check_response: true
    risk_level_bar: high
```

### Secret Management

Both plugins support APISIX secret management for credentials:

```yaml
plugins:
  ai-aws-content-moderation:
    comprehend:
      access_key_id: "$secret://vault/aws_key_id"
      secret_access_key: "$secret://vault/aws_secret_key"
      region: us-east-1
```

## Config Sync Example

```yaml
version: "1"
routes:
  - id: moderated-chat
    uri: /v1/chat/completions
    methods:
      - POST
    plugins:
      ai-aws-content-moderation:
        comprehend:
          access_key_id: AKIAIOSFODNN7EXAMPLE
          secret_access_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
          region: us-east-1
        moderation_categories:
          HATE_SPEECH: 0.3
          VIOLENCE_OR_THREAT: 0.2
        moderation_threshold: 0.5
      ai-proxy:
        provider: openai
        auth:
          header:
            Authorization: Bearer sk-your-key
        options:
          model: gpt-4
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| "no ai instance picked" | Aliyun plugin used without ai-proxy | Always configure ai-proxy or ai-proxy-multi on the same route |
| AWS plugin not blocking | Threshold too permissive | Lower `moderation_threshold` or per-category thresholds |
| Aliyun response moderation inactive | `check_response` defaults to `false` | Explicitly set `check_response: true` |
| "Specified signature is not matched" | Wrong Aliyun credentials | Verify `access_key_id` and `access_key_secret` |
| High latency | Double moderation (both plugins) | Use one moderation provider per route, not both |
| Streaming interrupted mid-response | Aliyun realtime mode detected violation | Expected behavior; adjust `risk_level_bar` or use `final_packet` mode |
