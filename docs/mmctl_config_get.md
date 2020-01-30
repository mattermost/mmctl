## mmctl config get

Get config setting

### Synopsis

Gets the value of a config setting by its name in dot notation.

```
mmctl config get [flags]
```

### Examples

```
config get SqlSettings.DriverName
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl config](mmctl_config.md)	 - Configuration

