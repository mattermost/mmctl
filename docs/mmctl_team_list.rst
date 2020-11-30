.. _mmctl_team_list:

mmctl team list
---------------

List all teams

Synopsis
~~~~~~~~


List all teams on the server.

::

  mmctl team list [flags]

Examples
~~~~~~~~

::

    team list

Options
~~~~~~~

::

  -h, --help   help for list

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

