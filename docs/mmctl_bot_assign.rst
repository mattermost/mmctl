.. _mmctl_bot_assign:

mmctl bot assign
----------------

Assign bot

Synopsis
~~~~~~~~


Assign the ownership of a bot to another user

::

  mmctl bot assign [bot-username] [new-owner-username] [flags]

Examples
~~~~~~~~

::

    bot assign testbot user2

Options
~~~~~~~

::

  -h, --help   help for assign

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

* `mmctl bot <mmctl_bot.rst>`_ 	 - Management of bots

