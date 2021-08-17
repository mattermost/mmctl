.. _mmctl_config_reload:

mmctl config reload
-------------------

Reload the server configuration

Synopsis
~~~~~~~~


Reload the server configuration in case you want to new settings to be applied.

::

  mmctl config reload [flags]

Examples
~~~~~~~~

::

  config reload

Options
~~~~~~~

::

  -h, --help   help for reload

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

