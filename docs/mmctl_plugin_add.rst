.. _mmctl_plugin_add:

mmctl plugin add
----------------

Add plugins

Synopsis
~~~~~~~~


Add plugins to your Mattermost server.

::

  mmctl plugin add [plugins] [flags]

Examples
~~~~~~~~

::

    plugin add hovercardexample.tar.gz pluginexample.tar.gz

Options
~~~~~~~

::

  -h, --help   help for add

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl plugin <mmctl_plugin.rst>`_ 	 - Management of plugins

