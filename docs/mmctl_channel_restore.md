## mmctl channel restore

Restore some channels

### Synopsis

Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

```
mmctl channel restore [channels] [flags]
```

### Examples

```
  channel restore myteam:mychannel
```

### Options

```
  -h, --help   help for restore
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl channel](mmctl_channel.md)	 - Management of channels

