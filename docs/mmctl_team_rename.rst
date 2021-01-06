.. _mmctl_team_rename:

mmctl team rename
-----------------

Rename team

Synopsis
~~~~~~~~


Rename an existing team

::

  mmctl team rename [team] [flags]

Examples
~~~~~~~~

::

    team rename old-team --display_name 'New Display Name'

Options
~~~~~~~

::

      --display_name string   Team Display Name
  -h, --help                  help for rename

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

