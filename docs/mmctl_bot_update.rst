.. _mmctl_bot_update:

mmctl bot update
----------------

Update bot

Synopsis
~~~~~~~~


Update bot information.

::

  mmctl bot update [username] [flags]

Examples
~~~~~~~~

::

    bot update testbot --username newbotusername

Options
~~~~~~~

::

      --description string    Optional. The new description text for the bot.
      --display-name string   Optional. The new display name for the bot.
  -h, --help                  help for update
      --username string       Optional. The new username for the bot.

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

* `mmctl bot <mmctl_bot.rst>`_ 	 - Management of bots

