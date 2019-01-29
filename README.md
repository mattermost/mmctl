# mmctl

A remote CLI tool for
[Mattermost](https://github.com/mattermost/mattermost-server): the
Open Source, self-hosted Slack-alternative.

## Install

To install the project in your `$GOPATH`, simply run:

```
go get -u github.com/mgdelacroix/mmctl
```

## Compile

First we have to install the dependencies of the project. `mmctl` uses
[Go dep](https://github.com/golang/dep) to manage the dependencies, so
after installing it, we need to run from the root of the project:

```sh
dep ensure
```

With the dependencies installed, we can compile the binary with:

```sh
go build -o bin/mmctl mmctl/main.go
```

## Usage

```sh
$ ./bin/mmctl
Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com

Usage:
  mmctl [command]

Available Commands:
  auth        Manages the credentials of the remote Mattermost instance
  channel     Management of channels
  help        Help about any command
  license     Licensing commands
  plugin      Management of plugins
  team        Management of teams
  user        Management of users

Flags:
  -h, --help   help for mmctl

Use "mmctl [command] --help" for more information about a command.
```

## Roadmap

 - [X] Login command
 - [X] Team command
 - [X] Channel command
 - [X] User command
 - [X] License command
 - [X] Plugin command
 - [ ] Config command
 - [ ] Command command
 - [ ] Roles command
 - [ ] Add more commands to the list
 - [ ] Unit tests
 - [ ] Credentials storage
 - [ ] Contexts
