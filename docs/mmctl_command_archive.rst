.. _mmctl_command_archive:

mmctl command archive
---------------------

Archive a slash command

Synopsis
~~~~~~~~


Archive a slash command. Commands can be specified by command ID.

::

  mmctl command archive [commandID] [flags]

Examples
~~~~~~~~

::

    command archive commandID

Options
~~~~~~~

::

  -h, --help   help for archive

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

* `mmctl command <mmctl_command.rst>`_ 	 - Management of slash commands

