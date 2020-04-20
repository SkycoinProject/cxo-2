# CXO File Sharing App

This is an example application built on top of the CXO 2.0 Node, for file sharing. The idea is that users can share files and directories among other users.

Each user is identified by pub key and must have a running cxo node instance. With the file sharing app users are able to publish new files and directories through the CXO Node. The CXO Node will notify all subscribed nodes with the new data hash so that each node can download published data and save new file structure locally. 

## Requirements

In order to run and use the file sharing app properly you must have CXO 2.0 Node installed and started.
For more info check [CXO 2.0 Node installation guide](/README.md).

## Installation

Run `sh example/cxo-file-sharing/install.sh` from the root of this project (_file sharing app and CLI will be installed in GOPATH/bin so make sure it's on your environment's PATH variable or use absolute path for both of them_).

## Structure

File sharing app consists of two parts:

1. Daemon service 
2. CLI

### Daemon Service

Service instance can be started with `cxo-file-sharing` command. Service has two key points implemented:

- App registration to CXO Node
- Listening and handling CXO Node notifications

#### App registration

App is registered on service startup. This is necessary in order to get notifications from cxo node every time when new data is retrieved.

The App is automatically registered by calling the CXO Node POST api `/api/v1/registerApp` and sending [model.RegisterAppRequest](/pkg/model/model.go). The address in the request should be the address of the API for notification handler (in this case `127.0.0.1:6430/notify`).

#### Listening to and handling CXO Node notifications

Service is listening to and handling notifications trough POST api `/notify` and accepting requests of the type [model.NotifyAppRequest](/pkg/model/model.go).
After the notification is accepted, request is processed and file structure is created in the desired location, in this case `$HOME/cxo-file-sharing`.

### CLI

CLI is used to interact with CXO 2.0 Node CLI. Usage: `cxo-file-sharing-cli publish <pathToFileOrFolder>`

After the CLI publish command is called, [model.PublishDataRequest](/pkg/model/model.go) is created by reading the file structure on the specified path.

Upon request creation, a temporary json file is created and CXO 2.0 CLI publish method is called with path to temporary file. This way the file structure is published through the CXO Node.