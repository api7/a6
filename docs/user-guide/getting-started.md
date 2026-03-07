# Getting Started

This guide helps you set up and use the a6 CLI to manage your Apache APISIX instance.

## Prerequisites

Before installing a6, ensure you have:

- Go 1.22 or higher.
- An Apache APISIX instance with the Admin API enabled. The default Admin API port is 9180.

## Installation

Install the a6 CLI directly with the Go command:

```bash
go install github.com/api7/a6/cmd/a6@latest
```

Alternatively, you can build from source:

```bash
git clone https://github.com/api7/a6.git
cd a6
make build
```

The resulting binary will be located in the current directory.

## Configuring Your First Context

The a6 CLI uses "contexts" to manage different APISIX environments. A context stores the server address and the Admin API key. Use the `context create` command to set up your first connection.

```bash
a6 context create local --server http://localhost:9180 --api-key edd1c9f034335f136f87ad84b625c8f1
```

Example output:

```bash
✓ Context "local" created and saved.
✓ Context "local" set as current context.
```

Your configuration is stored in `~/.config/a6/config.yaml` by default. You can override this location by setting the `A6_CONFIG_DIR` environment variable.

### Using Environment Variables

If you prefer not to use a context, you can set the following environment variables:

- `A6_SERVER`: The Admin API server address.
- `A6_API_KEY`: The Admin API key.

## Verifying the Connection

Check if a6 can communicate with your APISIX instance by listing the configured routes:

```bash
a6 route list
```

If the connection is successful, you will see a list of your existing routes.

## Your First Route

Once you have a working context, you can start managing routes.

### 1. Create a Route Configuration

Create a file named `route.json` with the following content:

```json
{
  "id": "getting-started",
  "name": "getting-started-route",
  "uri": "/get",
  "methods": ["GET"],
  "upstream": {
    "type": "roundrobin",
    "nodes": {
      "httpbin.org:80": 1
    }
  }
}
```

### 2. Apply the Route

Use the `route create` command to send this configuration to APISIX:

```bash
a6 route create -f route.json
```

### 3. Verify the Route

List your routes to see the new entry:

```bash
a6 route list
```

You can also get the full details of the route you just created:

```bash
a6 route get getting-started
```

### 4. Test the Route

Assuming your APISIX gateway is running and listening for data plane traffic (default port 9080), you can test the route with `curl`:

```bash
curl -i http://localhost:9080/get
```

### 5. Clean Up

When you are done, you can delete the route:

```bash
a6 route delete getting-started --force
```

## Managing Multiple Contexts

You can create multiple contexts for different environments like staging or production.

```bash
a6 context create staging --server http://staging.api:9180 --api-key YOUR_STAGING_KEY
```

To see all available contexts:

```bash
a6 context list
```

Example output:

```bash
NAME     SERVER                    CURRENT
local    http://localhost:9180      *
staging  http://staging.api:9180
```

To switch between contexts, use `context use`:

```bash
a6 context use staging
```

Example output:

```bash
✓ Switched to context "staging".
```

You can verify the active context anytime:

```bash
a6 context current
```

## What's Next

- Check the [Configuration Guide](configuration.md) for detailed configuration options.
- See the [Route Management Guide](route.md) for comprehensive route CRUD operations.
