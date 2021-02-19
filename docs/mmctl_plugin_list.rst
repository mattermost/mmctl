.. _mmctl_plugin_list:

mmctl plugin list
-----------------

List plugins

Synopsis
~~~~~~~~


List all enabled and disabled plugins installed on your Mattermost server.

::

  mmctl plugin list [flags]

Examples
~~~~~~~~

::

    plugin list

Options
~~~~~~~

::

  -h, --help   help for list

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

* `mmctl plugin <mmctl_plugin.rst>`_ 	 - Management of plugins

