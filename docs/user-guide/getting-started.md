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


## Interactive Mode

When you run a command that requires a resource ID without providing one,
a6 presents an interactive fuzzy-filterable list:

```bash
# Instead of remembering the route ID...
a6 route get

# a6 fetches all routes and presents a picker:
# > Select a route
#   my-api (1)
#   auth-service (2)
#   health-check (3)
```

This works for resource commands that support ID-based get, delete, update, and upstream health.
Interactive mode requires a terminal. In scripts or pipes, provide the ID explicitly.

## Debug Request Tracing

Use `a6 debug trace` to verify how a request flows through a route.

```bash
a6 debug trace getting-started
```

By default, this command:

- Fetches the route from the Admin API
- Sends a probe request to the APISIX gateway
- Shows response status, latency, and plugin execution details (when available)

Useful options:

```bash
a6 debug trace getting-started \
  --method GET \
  --path /get \
  --output json
```

You can also override data-plane and control-plane addresses when needed:

```bash
a6 debug trace getting-started \
  --gateway-url http://127.0.0.1:9080 \
  --control-url http://127.0.0.1:9090
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

## Shell Completion

The a6 CLI supports shell completion for bash, zsh, fish, and PowerShell.

### Bash

```bash
# Load completions in current session
source <(a6 completion bash)

# To load completions for every session (Linux)
a6 completion bash > /etc/bash_completion.d/a6

# To load completions for every session (macOS with Homebrew)
a6 completion bash > $(brew --prefix)/etc/bash_completion.d/a6
```

### Zsh

```bash
# Enable shell completion if not already done
echo "autoload -U compinit; compinit" >> ~/.zshrc

# Generate and install completion
a6 completion zsh > "${fpath[1]}/_a6"
```

You will need to start a new shell for this to take effect.

### Fish

```fish
# Load completions in current session
a6 completion fish | source

# To load completions for every session
a6 completion fish > ~/.config/fish/completions/a6.fish
```

### PowerShell

```powershell
# Load completions in current session
a6 completion powershell | Out-String | Invoke-Expression

# To load completions for every session, add to your profile
a6 completion powershell > a6.ps1
# Then source it from your PowerShell profile
```
