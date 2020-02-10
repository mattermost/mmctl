## mmctl user resetmfa

Turn off MFA

### Synopsis

Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they login.

```
mmctl user resetmfa [users] [flags]
```

### Examples

```
  user resetmfa user@example.com
```

### Options

```
  -h, --help   help for resetmfa
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl user](mmctl_user.md)	 - Management of users

