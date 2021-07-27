.. _mmctl_command_list:

mmctl command list
------------------

List all commands on specified teams.

Synopsis
~~~~~~~~


List all commands on specified teams.

::

  mmctl command list [teams] [flags]

Examples
~~~~~~~~

::

   command list myteam

Options
~~~~~~~

::

  -h, --help   help for list

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-file-path string      path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl command <mmctl_command.rst>`_ 	 - Management of slash commands

