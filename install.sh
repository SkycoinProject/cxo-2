#!/bin/bash

#TODO check this, failing on Ubuntu currently
#-set -e -o pipefail

go build -o "$GOPATH/bin/cxo-node" "$GOPATH/src/github.com/SkycoinProject/cxo-2/cmd/node"
go build -o "$GOPATH/bin/cxo-node-cli" "$GOPATH/src/github.com/SkycoinProject/cxo-2/cmd/cli"

#cli completion base directory
COMPLETION_DIR="$HOME/.cxo-node/.cli-completion"

[ -d "$COMPLETION_DIR" ] || mkdir -p "$COMPLETION_DIR"

BASH_FILE="$HOME/.bashrc" 
if [ -f "$BASH_FILE" ]; then
    echo "Installing cli completion for bash..."
    cp "$GOPATH/src/github.com/SkycoinProject/cxo-2/cmd/cli/bash_completion.sh" "$COMPLETION_DIR/cli-completion.bash"

    if ! grep -Fxq "source ~/.cxo-node/.cli-completion/cli-completion.bash" "$BASH_FILE" ; then
        echo "source ~/.cxo-node/.cli-completion/cli-completion.bash" >> "$BASH_FILE"
    fi
fi

ZSH_FILE="$HOME/.zshrc"
if [ -f "$ZSH_FILE" ]; then
    echo "Installing cli completion for zsh..."
    cp "$GOPATH/src/github.com/SkycoinProject/cxo-2/cmd/cli/zsh_completion.zsh" "$COMPLETION_DIR/_cxo-node-cli"

    if ! grep -Fxq "fpath=(~/.cxo-node/.cli-completion \$fpath)" "$ZSH_FILE" ; then
        echo "fpath=(~/.cxo-node/.cli-completion \$fpath)" >> "$ZSH_FILE"
        echo "autoload -Uz compinit && compinit" >> "$ZSH_FILE"
    fi
fi
