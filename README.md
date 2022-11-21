# mmctl

![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mmctl/master.svg)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mmctl/master.svg)](https://codecov.io/gh/mattermost/mmctl/branch/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattermost/mmctl)](https://goreportcard.com/report/github.com/mattermost/mmctl)

A remote CLI tool for [Mattermost](https://github.com/mattermost/mattermost-server): the Open Source, self-hosted Slack-alternative.

## Install

To install the project in your `$GOPATH`, simply run:

```
go install github.com/mattermost/mmctl/v6@master
```

### Install shell completions

To install the shell completions for bash, add the following line to your `~/.bashrc` or `~/.profile` file:

```sh
source <(mmctl completion bash)
```

For zsh, add the following line to file `~/.zshrc`:

```sh
source <(mmctl completion zsh)
```


## Usage

For the usage of all the commands, use the `--help` flag or check [the tool's documentation](./docs/mmctl.rst).

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
  post        Management of posts
  team        Management of teams
  user        Management of users
  websocket   Display websocket in a human-readable format

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

**NOTE:** `mmctl` is designed to run against a specific version of the `mattermost-server` and its API. If run against a server with a different version, `mmctl` will show a warning and will try to execute the commands. To ensure that the commands won't run if the server version is not supported, please use the `--strict` flag or set the `MMCTL_STRICT=true` environment variable.

## Login methods

### Password

```sh
$ mmctl auth login https://community.mattermost.com --name community --username my-username --password mysupersecret
```

The `login` command can also work interactively, so if you leave any needed flag empty, `mmctl` will ask you for it interactively:

```sh
$ mmctl auth login https://community.mattermost.com
Connection name: community
Username: my-username
Password:
```

### MFA

If you want to login with MFA, you just need to use the `--mfa-token` flag:

```sh
$ mmctl auth login https://community.mattermost.com --name community --username my-username --password mysupersecret --mfa-token 123456
```

### Access tokens

Instead of using username and password to log in, you can generate and use a personal access token to authenticate with a server:

```sh
$ mmctl auth login https://community.mattermost.com --name community --access-token MY_ACCESS_TOKEN
```


# Development

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/mattermost/mmctl)

## Compile

First we have to install the dependencies of the project. `mmctl` depends on go version 1.13.3 or greater.

We can compile the binary with:

```sh
make build
```

## Running the tests

`mmctl` has two types of tests: unit tests and end to end tests.

### Unit tests

To run the unit tests, you just need to execute:

```sh
make test
```

### End to end tests

To run the end to end test suite, you need to have the Mattermost server project downloaded and configured in your system. Check the [Developer Setup](https://developers.mattermost.com/contribute/server/developer-setup/) guide on how to configure a local server instance. The tests will search for a `mattermost-server` folder in the same directory where the `mmctl` is, but this can be changed through configuration if you have the project in a different path. Take a look at [the `config.mk` file](./config.mk) in this repository for instructions on how to make this change.

With the `mattermost-server` folder present, the only thing that needs to be done before running the tests themselves is to start the Mattermost docker development environment. The environment only needs to be started once, and then the tests can run as many times as needed. To start the docker environment, change to the `mattermost-server` project directory and run:

```sh
make start-docker
```

Change your directory back to `mmctl` and run the end to end test suite with:

```sh
make test-e2e
```
