# Sazed

Sazed is a small CLI tool to store and retrieve CLI commands.

## Usage

Sazed relies on a "Memories File", which has the following format:

```yaml
# File: memories.yaml
-----
- command: "echo foo"
  description: "Write foo to stdout"
- command: "ls -lha"
  description: "Lists files on current directory"
```

You can run sazed via CLI args or an env var

```yaml
# Via CLI Args
go run main.go --memories-file examples/memories.yaml

# Via ENV vargs
env SAZED_MEMORIES_FILE="./examples/memories.yaml" go run main.go
```
