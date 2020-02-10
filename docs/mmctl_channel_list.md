## mmctl channel list

List all channels on specified teams.

### Synopsis

List all channels on specified teams.
Archived channels are appended with ' (archived)'.

```
mmctl channel list [teams] [flags]
```

### Examples

```
  channel list myteam
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl channel](mmctl_channel.md)	 - Management of channels

