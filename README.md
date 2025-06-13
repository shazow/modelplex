# Modelplex

**Run AI agents in complete network isolation.** No outbound connections. Full AI capabilities.

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Guest VM      │    │   Modelplex      │    │   Providers     │
│                 │    │   (Host)         │    │                 │
│  ┌───────────┐  │    │  ┌─────────────┐ │    │ ┌─────────────┐ │
│  │    LLM    │  │    │  │   Model     │ │    │ │   OpenAI    │ │
│  │   Agent   │◄─┼────┼─►│ Multiplexer │◄┼────┼►│  Anthropic  │ │
│  └───────────┘  │    │  └─────────────┘ │    │ │   Ollama    │ │
│                 │    │  ┌─────────────┐ │    │ └─────────────┘ │
│                 │    │  │    MCP      │ │    │                 │
│                 │    │  │ Integration │◄┼────┼─► MCP Servers   │
│                 │    │  └─────────────┘ │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                        │
        └────modelplex.socket────┘
```

Your AI agent runs in a network-isolated VM, microcontainer, or air-gapped environment. It accesses GPT-4, Claude, Ollama, file systems, databases, and APIs through a single Unix socket. Modelplex handles all external communication from the host.

## Features

**🔀 Model Multiplexer**
- Use any model through one OpenAI-compatible interface: `gpt-4`, `claude-3-sonnet`, `llama2`
- Automatic failover with priority-based routing
- Environment variable API key management
- No API key management in isolated environment

**🛠️ MCP Tool Integration**  
- Access file tools, database connectors, APIs from isolation
- Model Context Protocol support for tool aggregation
- Fine-grained permission controls

**🔒 Zero Network Dependencies**
- Unix domain socket communication only
- Complete network isolation maintained
- No code changes to existing OpenAI-compatible agents

**📊 Full Observability**
- Structured logging with slog
- Monitor every AI interaction
- Request/response tracing
- Security vulnerability scanning

**🏗️ Production Ready**
- Comprehensive test suite with 55+ tests
- CI/CD with GitHub Actions
- Multi-platform builds (Linux, macOS, Windows)
- Docker support with minimal images
- Professional CLI with go-flags

## Use Cases

- **Secure Development**: Build AI agents in isolated environments
- **Enterprise Deployment**: AI in regulated, air-gapped networks  
- **Edge Computing**: Reliable AI without network dependencies
- **Research**: Study AI behavior in controlled environments
- **Multi-Provider Setup**: Seamlessly switch between OpenAI, Anthropic, and local models

## Architecture

### Core Components

- **Model Multiplexer**: Routes requests to providers based on model availability and priority
- **OpenAI-Compatible Proxy**: Full API compatibility with existing tools and libraries
- **Unix Socket Server**: HTTP server bound to Unix domain socket for isolation
- **Provider Abstraction**: Unified interface for OpenAI, Anthropic, Ollama, and future providers
- **MCP Integration**: Model Context Protocol support for tool aggregation

### Provider Support

| Provider | Authentication | Models | Notes |
|----------|---------------|---------|-------|
| **OpenAI** | Bearer token | GPT family | Full OpenAI API compatibility |
| **Anthropic** | x-api-key header | Claude family | Message format conversion, system message handling |
| **Ollama** | None | Any local model | Local inference, completion and chat endpoints |

### Security Features

- **Network Isolation**: Unix socket communication only
- **Secure Logging**: Structured logging with slog, no sensitive data exposure
- **Vulnerability Scanning**: CI/CD pipeline with govulncheck
- **Minimal Dependencies**: Only essential libraries, regularly updated

## Quick Start

### 1. Installation

```bash
# Build from source
git clone https://github.com/shazow/modelplex.git
cd modelplex
go build -o modelplex ./cmd/modelplex

# Or download binary from releases
```

### 2. Configuration

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
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
./modelplex --config config.toml --socket ./modelplex.socket --verbose
```

### 4. Use from Isolation

```python
# Isolated agent environment (Python)
import openai

client = openai.OpenAI(
    base_url="http://unix:/path/to/modelplex.socket",
    api_key="unused"  # Not needed, handled by host
)

# Works exactly like normal OpenAI, but completely isolated
response = client.chat.completions.create(
    model="gpt-4",  # Routes to provider with highest priority
    messages=[{"role": "user", "content": "Hello from isolation!"}]
)

print(response.choices[0].message.content)
```

```javascript
// Isolated agent environment (Node.js)
import OpenAI from 'openai';

const client = new OpenAI({
  baseURL: 'http://unix:/path/to/modelplex.socket',
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

## CLI Options

```bash
./modelplex --help

Usage:
  modelplex [OPTIONS]

Application Options:
  -c, --config=  Path to configuration file (default: config.toml)
  -s, --socket=  Path to Unix socket (default: ./modelplex.socket)
  -v, --verbose  Enable verbose logging
      --version  Show version information

Help Options:
  -h, --help     Show this help message
```


## Configuration

### Provider Configuration

Each provider supports different configuration options:

```toml
[[providers]]
name = "custom-openai"
type = "openai"
base_url = "https://api.openai.com/v1"  # Custom base URL for compatible APIs
api_key = "${OPENAI_API_KEY}"           # Environment variable substitution
models = ["gpt-4", "gpt-3.5-turbo"]     # Available models
priority = 1                            # Lower number = higher priority

[[providers]]
name = "anthropic"
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = "${ANTHROPIC_API_KEY}"
models = ["claude-3-sonnet", "claude-3-haiku", "claude-3-opus"]
priority = 2

[[providers]]
name = "local-llm"
type = "ollama"
base_url = "http://localhost:11434"     # Local Ollama instance
models = ["llama2", "codellama", "mistral"]
priority = 3                           # Fallback to local models
```

### MCP Configuration

```toml
[mcp]
enabled = true
servers = [
    { 
        name = "filesystem", 
        command = "mcp-server-filesystem", 
        args = ["/workspace", "--read-only"] 
    },
    { 
        name = "database", 
        command = "mcp-server-postgres", 
        args = ["postgresql://user:pass@localhost/db"] 
    }
]
```

### Environment Variables

- `OPENAI_API_KEY`: OpenAI API key
- `ANTHROPIC_API_KEY`: Anthropic API key
- `CONFIG_FILE`: Override default config file path
- `SOCKET_PATH`: Override default socket path

## Docker

```bash
# Build and run
docker build -t modelplex .
docker run -v /path/to/config.toml:/config.toml \
           -v /path/to/socket:/socket \
           modelplex --config /config.toml --socket /socket/modelplex.socket
```

## License

MIT License - see [LICENSE](LICENSE) file for details.
