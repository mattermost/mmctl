.. _mmctl_config_show:

mmctl config show
-----------------

Writes the server configuration to STDOUT

Synopsis
~~~~~~~~


Prints the server configuration and writes to STDOUT in JSON format.

::

  mmctl config show [flags]

Examples
~~~~~~~~

::

  config show

Options
~~~~~~~

::

  -h, --help   help for show

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

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

