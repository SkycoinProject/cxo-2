#compdef _cxo-node-cli cxo-node-cli


function _cxo-node-cli {
  local -a commands

  _arguments -C \
    "1: :->cmnds" \
    "*::arg:->args"

  case $state in
  cmnds)
    commands=(
      "publish:Publish new data to the CXO Tracker service"
      "subscribe:Subscribe to public key"
    )
    _describe "command" commands
    ;;
  esac

  case "$words[1]" in
  publish)
    _cxo-node-cli_publish
    ;;
  subscribe)
    _cxo-node-cli_subscribe
    ;;
  esac
}

function _cxo-node-cli_publish {
  _arguments
}

function _cxo-node-cli_subscribe {
  _arguments
}

