## mmctl auth login

Login into an instance

### Synopsis

Login into an instance and store credentials

```
mmctl auth login [instance url] --name [server name] --username [username] --password [password] [flags]
```

### Examples

```
  auth login https://mattermost.example.com
  auth login https://mattermost.example.com --name local-server --username sysadmin --password mysupersecret
  auth login https://mattermost.example.com --name local-server --username sysadmin --password mysupersecret --mfa-token 123456
  auth login https://mattermost.example.com --name local-server --access-token myaccesstoken
```

### Options

```
  -a, --access-token string   Access token to use instead of username/password
  -h, --help                  help for login
  -m, --mfa-token string      MFA token for the credentials
  -n, --name string           Name for the credentials
      --no-activate           If present, it won't activate the credentials after login
  -p, --password string       Password for the credentials
  -u, --username string       Username for the credentials
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
```

### SEE ALSO

* [mmctl auth](mmctl_auth.md)	 - Manages the credentials of the remote Mattermost instances

