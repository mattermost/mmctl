## mmctl command move

Move a slash command to a different team

### Synopsis

Move a slash command to a different team. Commands can be specified by command ID.

```
mmctl command move [team] [commandID] [flags]
```

### Examples

```
  command move newteam commandID
```

### Options

```
  -h, --help   help for move
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl command](mmctl_command.md)	 - Management of slash commands

