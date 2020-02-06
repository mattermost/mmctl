.. _mmctl_team_users_remove:

mmctl team users remove
-----------------------

Remove users from team

Synopsis
~~~~~~~~


Remove some users from team

::

  mmctl team users remove [team] [users] [flags]

Examples
~~~~~~~~

::

    team remove myteam user@example.com username

Options
~~~~~~~

::

  -h, --help   help for remove

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team users <mmctl_team_users.rst>`_ 	 - Management of team users

