.. _mmctl_plugin_enable:

mmctl plugin enable
-------------------

Enable plugins

Synopsis
~~~~~~~~


Enable plugins for use on your Mattermost server.

::

  mmctl plugin enable [plugins] [flags]

Examples
~~~~~~~~

::

    plugin enable hovercardexample pluginexample

Options
~~~~~~~

::

  -h, --help   help for enable

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl plugin <mmctl_plugin.rst>`_ 	 - Management of plugins

