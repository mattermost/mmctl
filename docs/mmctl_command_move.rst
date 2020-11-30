.. _mmctl_command_move:

mmctl command move
------------------

Move a slash command to a different team

Synopsis
~~~~~~~~


Move a slash command to a different team. Commands can be specified by command ID.

::

  mmctl command move [team] [commandID] [flags]

Examples
~~~~~~~~

::

    command move newteam commandID

Options
~~~~~~~

::

  -h, --help   help for move

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

* `mmctl command <mmctl_command.rst>`_ 	 - Management of slash commands

