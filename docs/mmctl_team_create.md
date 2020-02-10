## mmctl team create

Create a team

### Synopsis

Create a team.

```
mmctl team create [flags]
```

### Examples

```
  team create --name mynewteam --display_name "My New Team"
  team create --name private --display_name "My New Private Team" --private
```

### Options

```
      --display_name string   Team Display Name
      --email string          Administrator Email (anyone with this email is automatically a team admin)
  -h, --help                  help for create
      --name string           Team Name
      --private               Create a private team.
```

### Options inherited from parent commands

```
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one
```

### SEE ALSO

* [mmctl team](mmctl_team.md)	 - Management of teams

