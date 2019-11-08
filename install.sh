#!/bin/bash

go build -o "$GOPATH/bin/cxo-node" "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/node"
go build -o "$GOPATH/bin/cxo-node-cli" "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/cli"

[[ -d "$HOME/.cxo-node" ]] || mkdir "$HOME/.cxo-node"

if [ -f "$HOME/.bashrc" ]; then
    echo "Installing cli completion for bash..."
    cp "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/cli/bash_completion.sh" "$HOME/.cxo-node/.cli-completion.bash"

    if ! grep -Fxq "source ~/.cxo-node/.cli-completion.bash" "$HOME/.bashrc" ; then
        echo "source ~/.cxo-node/.cli-completion.bash" >> "$HOME/.bashrc"
    fi
fi

if [ -f "$HOME/.zshrc" ]; then
    echo "Installing cli completion for zsh..."
    cp "$GOPATH/src/github.com/SkycoinPro/cxo-2-node/cmd/cli/zsh_completion.zsh" "$HOME/.cxo-node/.cli-completion.zsh"

    if ! grep -Fxq "source ~/.cxo-node/.cli-completion.zsh" "$HOME/.zshrc" ; then
        echo "source ~/.cxo-node/.cli-completion.zsh" >> "$HOME/.zshrc"
    fi
fi
