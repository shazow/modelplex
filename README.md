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

[[providers]]
name = "anthropic"
type = "anthropic"
base_url = "https://api.anthropic.com/v1"
api_key = "${ANTHROPIC_API_KEY}"
models = ["claude-3-sonnet", "claude-3-haiku"]
priority = 2

[[providers]]
name = "local"
type = "ollama"
base_url = "http://localhost:11434"
models = ["llama2", "codellama"]
priority = 3

[mcp]
enabled = true
servers = [
    { name = "filesystem", command = "mcp-server-filesystem", args = ["/workspace"] },
    { name = "postgres", command = "mcp-server-postgres", args = ["postgresql://..."] }
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

## Development

### Building

```bash
# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run with coverage
go test -v -coverprofile=coverage.out ./...

# Run integration tests
go test -v -run Integration ./test/integration/...

# Build
go build -o modelplex ./cmd/modelplex
```

### Testing

The project includes comprehensive test coverage:

- **55+ tests** across all components
- **Unit tests** for providers, multiplexer, proxy, server
- **Integration tests** for full system validation
- **Mock-based testing** with testify framework
- **Race detection** for concurrent safety

```bash
# Run all tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...

# Generate coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### CI/CD

GitHub Actions pipeline includes:

- **Multi-Go version testing** (1.23, 1.24.4)
- **Linting** with golangci-lint
- **Security scanning** with gosec and govulncheck
- **Multi-platform builds** (Linux, macOS, Windows, ARM64)
- **Docker image building**
- **Codecov integration**

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

## API Compatibility

Modelplex provides full OpenAI API compatibility:

### Supported Endpoints

- `POST /v1/chat/completions` - Chat completions (all providers)
- `POST /v1/completions` - Text completions (OpenAI, Ollama)
- `GET /v1/models` - List available models
- `GET /health` - Health check endpoint

### Request/Response Format

All requests and responses follow OpenAI's specification. Provider-specific differences are handled internally:

- **Anthropic**: System messages moved to separate field, response format normalized
- **Ollama**: Stream parameter set to false, response format normalized
- **OpenAI**: Direct passthrough

## Docker

```dockerfile
# Multi-stage build for minimal image
FROM golang:1.24.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o modelplex ./cmd/modelplex

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/modelplex .
CMD ["./modelplex"]
```

```bash
# Build and run
docker build -t modelplex .
docker run -v /path/to/config.toml:/config.toml \
           -v /path/to/socket:/socket \
           modelplex --config /config.toml --socket /socket/modelplex.socket
```

## Roadmap

- [ ] **Real-time configuration updates** without restart
- [ ] **Advanced monitoring dashboard** with metrics and alerts  
- [ ] **Additional AI provider integrations** (Google AI, Azure OpenAI)
- [ ] **Enhanced MCP tool aggregation** with permission controls
- [ ] **WebSocket support** for streaming responses
- [ ] **Load balancing** across multiple provider instances
- [ ] **Request caching** and response optimization

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests (`go test -v ./...`)
4. Commit your changes (`git commit -am 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Create a Pull Request

### Development Guidelines

- **Go 1.24+** required
- **Comprehensive tests** for new features
- **Structured logging** with slog
- **Security-first** approach
- **OpenAI API compatibility** maintained

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Security

Report security vulnerabilities to [security@modelplex.dev](mailto:security@modelplex.dev).

- Regular security scanning with govulncheck
- Dependency vulnerability monitoring
- Structured logging prevents injection attacks
- Unix socket isolation for network security
