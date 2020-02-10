## mmctl license upload

Upload a license.

### Synopsis

Upload a license. Replaces current license.

```
mmctl license upload [license] [flags]
```

### Examples

```
  license upload /path/to/license/mylicensefile.mattermost-license
```

### Options

```
  -h, --help   help for upload
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl license](mmctl_license.md)	 - Licensing commands

