# Modelplex

Decouple models and MCP integrations from the agent.

**With modelplex, we can run AI agents in complete network isolation.** No outbound connections. Full AI capabilities.

```mermaid
graph LR
    subgraph GuestVM ["Guest VM"]
        LLMAgent["LLM Agent"]
    end
    
    subgraph Modelplex ["Modelplex (Host)"]
        ModelMux["Model Multiplexer"]
        MCPInt["MCP Integration"]
    end
    
    subgraph Providers ["Providers"]
        APIs["OpenAI<br/>Anthropic<br/>Ollama"]
        MCPServers["MCP Servers"]
    end
    
    LLMAgent <--> ModelMux
    ModelMux <--> APIs
    MCPInt <--> MCPServers
    
    GuestVM -.->|modelplex.socket| Modelplex
```

## Features

**🔀 Model Multiplexer**
- Use any model through one OpenAI-compatible interface
- Manage API keys and secrets in modelplex, so your agent doesn't need to know about them.

**🔒 Zero Network Dependencies**
- Unix domain socket communication only
- Run agents in a VM without a network device!

**📊 Full Observability**
- Structured logging with slog
- Monitor every AI interaction

## Quick Start

```bash
# Build from source
git clone https://github.com/shazow/modelplex.git
cd modelplex
go build -o modelplex ./cmd/modelplex
```

Create `config.toml`:

```toml
# Multi-provider configuration with failover
[[providers]]
name = "openai"
type = "openai"
base_url = "https://api.openai.com/v1"
api_key = "${OPENAI_API_KEY}"
models = ["gpt-4", "gpt-3.5-turbo"]
priority = 1

[mcp]
enabled = true
servers = [
    { name = "filesystem", command = "mcp-server-filesystem", args = ["/workspace"] },
]
```

### 3. Start Modelplex

```bash
# Host system (with network access)
./modelplex --config config.toml --socket ./modelplex.socket --verbose
```

### 4. Connect with an agent

```python
# Isolated agent environment (Python)
import openai

client = openai.OpenAI(
    base_url="unix:/path/to/modelplex.socket",
    api_key="unused"  # Not needed, handled by host
)

# Works exactly like normal OpenAI, but completely isolated
response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello from isolation!"}]
)

print(response.choices[0].message.content)
```

```javascript
// Isolated agent environment (Node.js)
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'unix:/path/to/modelplex.socket',
  apiKey: 'unused'
});

const response = await client.chat.completions.create({
  model: 'claude-3-sonnet',
  messages: [{ role: 'user', content: 'Hello from isolation!' }]
});
```

```bash
# Isolated agent environment (curl)
curl --unix-socket ./modelplex.socket \
  -X POST http://localhost/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```


## Docker

```bash
# Build the image
docker build -t modelplex .

# Run with custom config and socket directory
docker run -v /path/to/your/config.toml:/app/config.toml \
           -v /path/to/socket/dir:/tmp/modelplex \
           modelplex

# Or run with default config and expose socket to host
docker run -v /host/socket/dir:/tmp/modelplex \
           modelplex
```

The socket will be available at `/host/socket/dir/modelplex.socket` on your host system.

## License

MIT
