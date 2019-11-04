# cxo-2

In order to run the node following steps should be completed:
1/ obtain CXO Tracker's pub key value folloving the steps given on that project's README.
2/ update local pointer to the tracker service with that pub key
3/ run `sh install.sh` from the root of this project (node and CLI will be installed in GOPATH/bin so make sure it's on your environment's PATH variable or use absolute path for both of them)

Node instance is available running the `cxo-node`. That will run daemon service on user's local machine.
Node CLI is available with running the `cxo-node-cli`. It enables user to interact with CXO node for example by subscribing to pub key or announcing new content on the CXO.
