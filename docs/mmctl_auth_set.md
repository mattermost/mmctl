## mmctl auth set

Set the credentials to use

### Synopsis

Set an credentials to use in the following commands

```
mmctl auth set [server name] [flags]
```

### Examples

```
  auth set local-server
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl auth](mmctl_auth.md)	 - Manages the credentials of the remote Mattermost instances

