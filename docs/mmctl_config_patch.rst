.. _mmctl_config_patch:

mmctl config patch
------------------

Patch the config

Synopsis
~~~~~~~~


Patches config settings with the given config file.

::

  mmctl config patch <config-file> [flags]

Examples
~~~~~~~~

::

  config patch /path/to/config.json

Options
~~~~~~~

::

  -h, --help   help for patch

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

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

