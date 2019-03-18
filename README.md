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
go modules to manage the dependencies, so you need to have installed
go 1.11 or greater.

Dependencies will be managed by go automatically, but if you want to
install them locally on the vendor folder, just run:

```sh
make vendor
```

We can compile the binary with:

```sh
make build
```

## Usage

```sh
Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com

Usage:
  mmctl [command]

Available Commands:
  auth        Manages the credentials of the remote Mattermost instances
  channel     Management of channels
  help        Help about any command
  license     Licensing commands
  permissions Management of permissions and roles
  plugin      Management of plugins
  team        Management of teams
  user        Management of users

Flags:
  -h, --help   help for mmctl

Use "mmctl [command] --help" for more information about a command.
```

First we have to log into a mattermost instance:

```sh
$ mmctl auth login my-instance https://my-instance.example.com john.doe mysupersecret

  credentials for my-instance: john.doe@https://my-instance.example.com stored

```

We can check the currently stored credentials with:

```sh
$ mmctl auth list

    | Active |        Name | Username |                     InstanceUrl |
    |--------|-------------|----------|---------------------------------|
    |      * | my-instance | john.doe | https://my-instance.example.com |

```

And now we can run commands normaly:

```sh
$ mmctl user search john.doe
id: qykfw3t933y38k57ubct77iu9c
username: john.doe
nickname:
position:
first_name: John
last_name: Doe
email: john.doe@example.com
auth_service:
```

## Roadmap

 - [X] Login command
 - [X] Team command
 - [X] Channel command
 - [X] User command
 - [X] License command
 - [X] Plugin command
 - [X] Authentication Contexts
 - [ ] Command command
 - [ ] Config command
 - [ ] ? Export command
 - [ ] ? Import command
 - [ ] ldap command
 - [ ] ? logs command
 - [ ] roles command
 - [ ] ? version command
 - [ ] Unit tests
 - [ ] Autocomplete
