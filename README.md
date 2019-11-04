# CXO 2.0 Node & CLI

This repo contains the CXO 2.0 Node and CLI. These tools allow you to make use of the CXO 2.0 Tracker.

## Installation

In order to run the Node or CLI, the following steps should be completed (_skip 1 and 2 if you are using the default cxo.skycoin.com tracker_):
1. Obtain CXO Tracker's pub key value following the steps given on that project's README (_once public, this should be published somewhere else_)
2. Update local pointer to the tracker service with that pub key
3. Run `sh install.sh` from the root of this project (_node and CLI will be installed in GOPATH/bin so make sure it's on your environment's PATH variable or use absolute path for both of them_)

## CXO 2.0 Node

The Node handles all background activity and communication with the CXO 2.0 Tracker. This includes:
- Listening to subscribed data feed updates
- Requesting and downloading newly published files from subscribed data feeds

The Node is purely a background service and does not require any commands for usage. 

Node instance is available running the `cxo-node`. Executing it will run the daemon service on the user's local machine. 

### Database

The Node stores data locally in a BoltDB bucket called Data Objects. Currently it's storing the data object hash and the path to the local file. This is being improved further at the moment.

## CXO 2.0 CLI

The CLI may be used manually or called upon from other applications. The CLI is available by running the `cxo-node-cli`. It enables users to interact with the CXO 2.0 Tracker and allows:

- Subscribing to pub key
    Example usage:
    `cxo-node-cli subscribe <publisher's pub key>`
- Publishing new objects (_includes signing of the object_)
    Example usage:
    `cxo-node-cli publish <pathToFile>`
