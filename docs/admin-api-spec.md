# APISIX Admin API Specification

Extracted from Apache APISIX v3.x source code. This is the authoritative reference for a6 CLI code generation.

## Base Configuration

- Default port: 9180
- API prefix: /apisix/admin
- Protocol: HTTP/HTTPS

## Authentication

- Header: X-API-KEY
- Alternative: query parameter `api_key` or cookie `x_api_key`
- Roles:
  - `admin`: Full access
  - `viewer`: GET access only

## Response Format

### Single Resource
```json
{
  "key": "/apisix/routes/1",
  "value": {...},
  "createdIndex": N,
  "modifiedIndex": N
}
```

### List of Resources
```json
{
  "total": N,
  "list": [
    {
      "key": "/apisix/routes/1",
      "value": {...},
      "createdIndex": N,
      "modifiedIndex": N
    },
    ...
  ]
}
```

## Resources

### 1. Route (/apisix/admin/routes)

Routes match client requests based on defined rules, execute plugins, and forward requests to upstreams.

- **Methods**:
  - `GET /apisix/admin/routes`: Fetches a list of all configured Routes.
  - `GET /apisix/admin/routes/:id`: Fetches specified Route by ID.
  - `PUT /apisix/admin/routes/:id`: Creates or updates a Route with the specified ID.
  - `POST /apisix/admin/routes`: Creates a Route and assigns a random ID.
  - `PATCH /apisix/admin/routes/:id`: Standard PATCH to modify specified attributes.
  - `PATCH /apisix/admin/routes/:id/{path}`: Subpath PATCH to update a specific attribute.
  - `DELETE /apisix/admin/routes/:id`: Removes the Route.

- **Query Parameters**:
  - `page` (integer): Page number for list results (default: 1).
  - `page_size` (integer): Number of resources per page (10-500, default: 10).
  - `name` (string): Filter by Route name.
  - `label` (string): Filter by label.
  - `uri` (string): Filter by URI.
  - `filter` (string): URL-encoded filter (e.g., `service_id=1`).
  - `ttl` (integer): Set auto-expiry in seconds (supported by PUT).
  - `force` (boolean): Force delete if `true`.

- **Schema Fields**:
  - `id` (string|integer): Identifier. Max 64 chars for string. Pattern: `^[a-zA-Z0-9-_.]+$`.
  - `uri` (string): Matches the URL. Required if `uris` not set. Supports path parameters via `{var}`.
  - `uris` (array of strings): Non-empty list of URIs. Required if `uri` not set.
  - `name` (string): Max 256 chars.
  - `desc` (string): Max 256 chars.
  - `host` (string): Matches domain names (e.g., `foo.com` or `*.foo.com`). Mutually exclusive with `hosts`.
  - `hosts` (array of strings): Non-empty list of hosts. Mutually exclusive with `host`.
  - `remote_addr` (string): Matches client IP (IPv4, CIDR, or IPv6). Mutually exclusive with `remote_addrs`.
  - `remote_addrs` (array of strings): Non-empty list of remote addresses. Mutually exclusive with `remote_addr`.
  - `methods` (array of strings): Supported: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`, `OPTIONS`, `CONNECT`, `TRACE`, `PURGE`.
  - `priority` (integer): Match priority. Higher value means higher priority. Default: 0.
  - `vars` (array): List of `[variable, operator, value]` match rules.
  - `filter_func` (string): Lua function for custom filtering. Must start with `function`.
  - `plugins` (object): Key-value pairs of plugin names and their configurations.
  - `plugin_config_id` (string|integer): Reference to a Plugin Config resource. Mutually exclusive with `script`.
  - `upstream` (object): Inline Upstream definition.
  - `upstream_id` (string|integer): Reference to an Upstream resource.
  - `service_id` (string|integer): Reference to a Service resource.
  - `labels` (object): Key-value pairs. Keys/values max 256 chars. Pattern: `^\S+$`.
  - `timeout` (object): Contains `connect`, `send`, `read` (numbers, in seconds).
  - `enable_websocket` (boolean): Enables websocket. Default: false.
  - `status` (integer): `1` (enabled) or `0` (disabled). Default: 1.
  - `script` (string): Lua code for plugin orchestration. Max 100KB.
  - `create_time` (integer): Epoch timestamp (read-only).
  - `update_time` (integer): Epoch timestamp (read-only).

### 2. Service (/apisix/admin/services)

A Service is an abstraction of an API, usually corresponding to an upstream service.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `POST`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`, `name`, `label`
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `name` (string): Identifier for the Service.
  - `desc` (string): Description.
  - `plugins` (object): Bound plugins.
  - `upstream` (object): Inline Upstream configuration.
  - `upstream_id` (string|integer): Reference to Upstream.
  - `labels` (object): Key-value pairs.
  - `enable_websocket` (boolean): Default: false.
  - `hosts` (array of strings): Non-empty list of hosts.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 3. Upstream (/apisix/admin/upstreams)

Upstreams perform load balancing on a set of service nodes.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `POST`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`, `name`, `label`, `force`
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `name` (string): Identifier for the Upstream.
  - `desc` (string): Description.
  - `type` (string): Load balancing algorithm: `roundrobin`, `chash`, `ewma`, `least_conn`. Default: `roundrobin`.
  - `nodes` (object or array):
    - Object: `{"host:port": weight, ...}`
    - Array: `[{"host": "...", "port": ..., "weight": ..., "priority": ..., "metadata": {...}}, ...]`
  - `service_name` (string): For service discovery. Required if nodes not set.
  - `discovery_type` (string): For service discovery. Required if `service_name` set.
  - `hash_on` (string): For `chash` type: `vars`, `header`, `cookie`, `consumer`, `vars_combinations`. Default: `vars`.
  - `key` (string): Hash key used with `hash_on`.
  - `checks` (object): Health check configuration (active and passive).
  - `retries` (integer): Number of retries.
  - `retry_timeout` (number): Timeout for retries.
  - `timeout` (object): `{connect, send, read}` in seconds.
  - `pass_host` (string): `pass`, `node`, `rewrite`. Default: `pass`.
  - `upstream_host` (string): Used when `pass_host` is `rewrite`.
  - `scheme` (string): `http`, `https`, `grpc`, `grpcs`, `tcp`, `tls`, `udp`, `kafka`. Default: `http`.
  - `labels` (object): Key-value pairs.
  - `keepalive_pool` (object): `{size, idle_timeout, requests}`.
  - `tls` (object): `{client_cert, client_key, client_cert_id, verify}`.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 4. Consumer (/apisix/admin/consumers)

Consumers represent users of services. Identified by `username`.

- **Methods**: `GET` (list), `GET /:username`, `PUT`, `DELETE /:username`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `username` (string): Required, unique. Pattern: `^[a-zA-Z0-9_\-]+$`.
  - `plugins` (object): Authentication and other plugins.
  - `group_id` (string|integer): Reference to a Consumer Group.
  - `desc` (string): Description.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 5. Consumer Credential (/apisix/admin/consumers/:username/credentials)

Stored credentials for a specific Consumer.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `DELETE /:id`
- **Path Parameter**: `:username` (Consumer name)
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `plugins` (object): Configuration for ONE authentication plugin.
  - `name` (string): Identifier.
  - `desc` (string): Description.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 6. SSL (/apisix/admin/ssls)

SSL/TLS certificates for SNI matching or client authentication.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `POST`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `cert` (string): HTTPS certificate (PEM). Supports Secret Manager URIs. Required.
  - `key` (string): HTTPS private key (PEM). Supports Secret Manager URIs. Required.
  - `certs` (array of strings): Additional certificates for the same domain.
  - `keys` (array of strings): Additional private keys pairing with `certs`.
  - `sni` (string): Server Name Indication. Required if `snis` not set and `type` is `server`.
  - `snis` (array of strings): List of SNIs. Required if `sni` not set and `type` is `server`.
  - `client` (object): `{ca, depth, skip_mtls_uri_regex}` for mTLS.
  - `type` (string): `server` (default) or `client` (for upstream access).
  - `status` (integer): `1` (enabled) or `0` (disabled). Default: 1.
  - `ssl_protocols` (array of strings): Supported: `TLSv1.1`, `TLSv1.2`, `TLSv1.3`.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 7. Global Rule (/apisix/admin/global_rules)

Plugins that run globally before any Route or Service level plugins.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier. Required for PUT.
  - `plugins` (object): Plugins to run globally. Required.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 8. Plugin Config (/apisix/admin/plugin_configs)

Reusable sets of plugin configurations.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier. Required.
  - `name` (string): Identifier for the config.
  - `desc` (string): Description.
  - `plugins` (object): Plugins configuration. Required.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 9. Plugin Metadata (/apisix/admin/plugin_metadata)

Metadata for specific plugins. Keyed by plugin name.

- **Methods**: `GET /:plugin_name`, `PUT /:plugin_name`, `DELETE /:plugin_name`
- **Schema Fields**: Defined according to each plugin's `metadata_schema`.

### 10. Plugin (/apisix/admin/plugins)

Operations related to plugin information and control.

- **Methods**:
  - `GET /apisix/admin/plugins/list`: Fetches a list of all plugin names.
  - `GET /apisix/admin/plugins/:plugin_name`: Fetches plugin schema and attributes.
  - `GET /apisix/admin/plugins?all=true`: Fetches properties of all plugins.
  - `PUT /apisix/admin/plugins/reload`: Hot-reloads plugins from code changes.
- **Query Parameters**:
  - `subsystem`: `http` (default) or `stream`.

### 11. Stream Route (/apisix/admin/stream_routes)

Routes for L4 (TCP/UDP) traffic.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `POST`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `name` (string): Identifier.
  - `desc` (string): Description.
  - `remote_addr` (string): Matches client IP.
  - `server_addr` (string): Matches APISIX server IP.
  - `server_port` (integer): Matches APISIX server port.
  - `sni` (string): Server Name Indication.
  - `upstream` (object): Inline Upstream configuration.
  - `upstream_id` (string|integer): Reference to Upstream.
  - `service_id` (string|integer): Reference to Service.
  - `plugins` (object): L4 plugins.
  - `protocol` (object): xRPC protocol config `{name, superior_id, conf, logger}`.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 12. Proto (/apisix/admin/protos)

Storage for Protocol Buffer definitions used in gRPC transcoding.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `POST`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier.
  - `content` (string): Protobuf definition string. Required.
  - `name` (string): Identifier.
  - `desc` (string): Description.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 13. Consumer Group (/apisix/admin/consumer_groups)

Reusable sets of plugins for Consumers.

- **Methods**: `GET` (list), `GET /:id`, `PUT /:id`, `PATCH /:id`, `DELETE /:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields**:
  - `id` (string|integer): Identifier. Required.
  - `name` (string): Identifier.
  - `desc` (string): Description.
  - `plugins` (object): Plugins configuration. Required.
  - `labels` (object): Key-value pairs.
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

### 14. Secret (/apisix/admin/secrets)

Configuration for secret managers (e.g., Vault, AWS, GCP).

- **Methods**: `GET` (list), `GET /:manager/:id`, `PUT /:manager/:id`, `PATCH /:manager/:id`, `DELETE /:manager/:id`
- **Query Parameters**: `page`, `page_size`
- **Schema Fields (Vault)**:
  - `uri` (string): Vault server URI. Required.
  - `prefix` (string): Key prefix. Required.
  - `token` (string): Vault token. Required.
  - `namespace` (string): Vault namespace.
- **Schema Fields (AWS)**:
  - `access_key_id` (string): Required.
  - `secret_access_key` (string): Required.
  - `region` (string): AWS region.
  - `endpoint_url` (string): Custom endpoint URL.
- **Schema Fields (GCP)**:
  - `auth_config` (object): `{client_email, private_key, project_id, ...}`.
  - `auth_file` (string): Path to JSON auth file.
- **Common Fields**:
  - `id` (string): Format `:manager/:id` (e.g., `vault/1`).
  - `create_time` (integer): Read-only timestamp.
  - `update_time` (integer): Read-only timestamp.

## Schema Endpoints

Special endpoints to retrieve JSON schemas for resources and plugins.

- `GET /apisix/admin/schema/plugins/:plugin_name`: Returns JSON schema of a specific plugin.
- `GET /apisix/admin/schema/:resource_name`: Returns JSON schema of a resource type (e.g., `routes`, `upstreams`).
- `POST /apisix/admin/schema/validate/:resource_name`: Validates a resource configuration against its schema without saving.
