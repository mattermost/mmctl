## mmctl config set

Set config setting

### Synopsis

Sets the value of a config setting by its name in dot notation. Accepts multiple values for array settings

```
mmctl config set [flags]
```

### Examples

```
config set SqlSettings.DriverName mysql
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

* [mmctl config](mmctl_config.md)	 - Configuration

