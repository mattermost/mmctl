.. _mmctl_team_restore:

mmctl team restore
------------------

Restore teams

Synopsis
~~~~~~~~


Restores archived teams.

::

  mmctl team restore [teams] [flags]

Examples
~~~~~~~~

::

    team restore myteam

Options
~~~~~~~

::

  -h, --help   help for restore

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

