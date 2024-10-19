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

## Installing

### Binary

Git clone the repo and run `make install`. Use `DIR` to customize the directory. If not given defaults to `DIR=~/.local/bin`.

```sh
make install
make install DIR=/bin
```

### Memories File

Create a memories file in `~/.config/sazed/memories.yaml`. For example:

```yaml
# file: ~/.config/sazed/memories.yaml
- command: 'docker ps -a -q | xargs -I{} docker stop {} && docker ps -a -q | xargs -I{} docker rm {}'
  description: Stop all running docker containers

- command: "git config --local user.email 'myemail@gmail.com' && git config --local user.name 'Some User'"
  description: Configures git to Some User

- command: find -iname '*.<ext>'
  description: Finds all files with common extension

- command: sudo systemctl restart iwd
  description: Restarts iwd
```

## Hooking to Bash

You can run sazed from a Bash with a shortcut by adding something like this to your .bashrc:
```sh
# ------------------------------------------------------------
# Sazed
# ------------------------------------------------------------
__sazed_run() {
    if ! command -v sazed 2>&1 >/dev/null
    then
        return 0
    fi
    local output
    output=$(sazed $1)
    READLINE_LINE=${output}
    READLINE_POINT=0x7fffffff
}

# Run for a local `.memories.yaml`.
bind -m emacs-standard -x '"\C-n": __sazed_run --memories-file=.memories.yaml'
bind -m vi-command -x '"\C-n": __sazed_run --memories-file=.memories.yaml'
bind -m vi-insert -x '"\C-n": __sazed_run --memories-file=.memories.yaml'

# Run for a global `.memories.yaml`
bind -m emacs-standard -x '"\en": __sazed_run'
bind -m vi-command -x '"\en": __sazed_run'
bind -m vi-insert -x '"\en": __sazed_run'
```
