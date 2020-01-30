## mmctl channel rename

Rename channel

### Synopsis

Rename an existing channel.

```
mmctl channel rename [channel] [flags]
```

### Examples

```
  channel rename myteam:oldchannel --name 'new-channel' --display_name 'New Display Name'
  channel rename myteam:oldchannel --name 'new-channel'
  channel rename myteam:oldchannel --display_name 'New Display Name'
```

### Options

```
      --display_name string   Channel Display Name
  -h, --help                  help for rename
      --name string           Channel Name
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl channel](mmctl_channel.md)	 - Management of channels

