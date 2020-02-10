## mmctl channel remove

Remove users from channel

### Synopsis

Remove some users from channel

```
mmctl channel remove [channel] [users] [flags]
```

### Examples

```
  channel remove myteam:mychannel user@example.com username
  channel remove myteam:mychannel --all-users
```

### Options

```
      --all-users   Remove all users from the indicated channel.
  -h, --help        help for remove
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl channel](mmctl_channel.md)	 - Management of channels

