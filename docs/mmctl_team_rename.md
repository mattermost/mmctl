## mmctl team rename

Rename team

### Synopsis

Rename an existing team

```
mmctl team rename [team] [flags]
```

### Examples

```
 team rename myoldteam newteamname --display_name 'New Team Name'
	team rename myoldteam - --display_name 'New Team Name'
```

### Options

```
      --display_name string   Team Display Name
  -h, --help                  help for rename
```

### Options inherited from parent commands

```
      --format string   the format of the command output [plain, json] (default "plain")
```

### SEE ALSO

* [mmctl team](mmctl_team.md)	 - Management of teams

