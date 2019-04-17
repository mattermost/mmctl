# mmctl

A remote CLI tool for
[Mattermost](https://github.com/mattermost/mattermost-server): the
Open Source, self-hosted Slack-alternative.

## Install

To install the project in your `$GOPATH`, simply run:

```
go get -u github.com/mgdelacroix/mmctl
```

### Install shell completions

To install the shell completions for bash, add the following line to
your `~/.bashrc` or `~/.profile` file:

```sh
source <(mmctl completion bash)
```

For zsh, add the following line to file `~/.zshrc`:

```sh
source <(mmctl completion zsh)
```

## Compile

First we have to install the dependencies of the project. `mmctl` uses
go modules to manage the dependencies, so you need to have installed
go 1.11 or greater.

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
  completion  Generates autocompletion scripts for bash and zsh
  group       Management of groups
  help        Help about any command
  license     Licensing commands
  logs        Display logs in a human-readable format
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
$ mmctl auth login https://my-instance.example.com --name my-instance --username john.doe --password mysupersecret

  credentials for my-instance: john.doe@https://my-instance.example.com stored

```

We can check the currently stored credentials with:

```sh
$ mmctl auth list

    | Active |        Name | Username |                     InstanceUrl |
    |--------|-------------|----------|---------------------------------|
    |      * | my-instance | john.doe | https://my-instance.example.com |

```

And now we can run commands normally:

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

## Login methods

### Password

```sh
$ mmctl auth login https://community.mattermost.com --name community --username my-username --password mysupersecret
```

The `login` command can also work interactively, so if you leave any
needed flag empty, `mmctl` will ask you for it interactively:

```sh
$ mmctl auth login https://community.mattermost.com
Connection name: community
Username: my-username
Password:
```

### MFA

If you want to login with MFA, you just need to use the `--mfa-token`
flag:

```sh
$ mmctl auth login https://community.mattermost.com --name community --username my-username --password mysupersecret --mfa-token 123456
```

### Access tokens

Instead of using username and password to log in, you can generate and
use a personal access token to authenticate with a server:

```sh
$ mmctl auth login https://community.mattermost.com --name community --username my-username --access-token MY_ACCESS_TOKEN
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
 - [ ] ldap command
 - [X] logs command
 - [ ] roles command
 - [X] Shell completions
