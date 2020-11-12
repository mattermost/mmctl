.. _mmctl_config_get:

mmctl config get
----------------

Get config setting

Synopsis
~~~~~~~~


Gets the value of a config setting by its name in dot notation.

::

  mmctl config get [flags]

Examples
~~~~~~~~

::

  config get SqlSettings.DriverName

Options
~~~~~~~

::

  -h, --help   help for get

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

