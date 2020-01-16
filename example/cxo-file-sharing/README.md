# CXO File Sharing App

This is example application built on top of CXO 2.0 Node for file sharing.

## Requirements

In order to run and use file sharing app properly you must have CXO 2.0 Node installed and started.
For more info check [CXO 2.0 Node installation guide](README.md).

## Structure

File sharing app consists from two parts:

1. Daemon service 
2. CLI

### Daemon Service

Service has two key points implemented

- App registration to cxo node
- Listening and handling cxo node notifications

#### App registration

App is registered on service startup. This is necessary in order to get notifications from cxo node every time when new data is retrieved.
App is registered by calling cxo node POST api `/api/v1/registerApp` and sending [model.RegisterAppRequest](/pkg/model/model.go).
Address in request should be address of api for notification handler in this case `127.0.0.1:8080/notify`.

#### Listening and handling cxo node notifications

Service is listening and handling notifications trough POST api `/notify` and accepting request type [model.NotifyAppRequest](/pkg/model/model.go).
After notification is accepted, request is processed and file structure is created in desired location, in this case `user-home-dir/cxo-file-sharing`.

### CLI

CLI is used to interact with CXO 2.0 Node CLI. Usage: `cxo-file-sharing-cli publish <pathToFileOrFolder>`

After cli publish command is called, [model.PublishDataRequest](/pkg/model/model.go) is assembled by scanning and digesting file structure on specified path.
Upon request creation, temporary json file is created and CXO 2.0 CLI publish method is called with path to temporary file. In this way file structure is published trough cxo node.







