.. _mmctl_command_show:

mmctl command show
------------------

Show a custom slash command

Synopsis
~~~~~~~~


Show a custom slash command. Commands can be specified by command ID. Returns command ID, team ID, trigger word, display name and creator username.

::

  mmctl command show [commandID] [flags]

Examples
~~~~~~~~

::

    command show commandID

Options
~~~~~~~

::

  -h, --help   help for show

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl command <mmctl_command.rst>`_ 	 - Management of slash commands

