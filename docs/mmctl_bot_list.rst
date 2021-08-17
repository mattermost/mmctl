.. _mmctl_bot_list:

mmctl bot list
--------------

List bots

Synopsis
~~~~~~~~


List the bots users.

::

  mmctl bot list [flags]

Examples
~~~~~~~~

::

    bot list

Options
~~~~~~~

::

      --all        Optional. Show all bots (including deleleted and orphaned).
  -h, --help       help for list
      --orphaned   Optional. Only show orphaned bots.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --config-path string           path to the configuration directory. (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl bot <mmctl_bot.rst>`_ 	 - Management of bots

