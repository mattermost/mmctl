## mmctl plugin disable

Disable plugins

### Synopsis

Disable plugins. Disabled plugins are immediately removed from the user interface and logged out of all sessions.

```
mmctl plugin disable [plugins] [flags]
```

### Examples

```
  plugin disable hovercardexample pluginexample
```

### Options

```
  -h, --help   help for disable
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl plugin](mmctl_plugin.md)	 - Management of plugins

