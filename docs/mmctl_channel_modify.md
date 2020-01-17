## mmctl channel modify

Modify a channel's public/private type

### Synopsis

Change the public/private type of a channel.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

```
mmctl channel modify [channel] [flags]
```

### Examples

```
  channel modify myteam:mychannel --private
```

### Options

```
  -h, --help      help for modify
      --private   Convert the channel to a private channel
      --public    Convert the channel to a public channel
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
```

### SEE ALSO

* [mmctl channel](mmctl_channel.md)	 - Management of channels

