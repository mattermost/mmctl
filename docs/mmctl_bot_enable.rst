.. _mmctl_bot_enable:

mmctl bot enable
----------------

Enable bot

Synopsis
~~~~~~~~


Enable a disabled bot

::

  mmctl bot enable [username] [flags]

Examples
~~~~~~~~

::

    bot enable testbot

Options
~~~~~~~

::

  -h, --help   help for enable

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

* `mmctl bot <mmctl_bot.rst>`_ 	 - Management of bots

