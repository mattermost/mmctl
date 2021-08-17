.. _mmctl_plugin_marketplace_install:

mmctl plugin marketplace install
--------------------------------

Install a plugin from the marketplace

Synopsis
~~~~~~~~


Installs a plugin listed in the marketplace server

::

  mmctl plugin marketplace install <id> [version] [flags]

Examples
~~~~~~~~

::

    # you can specify with both the plugin id and its version
    $ mmctl plugin marketplace install jitsi 2.0.0

    # if you don't specify the version, the latest one will be installed
    $ mmctl plugin marketplace install jitsi

Options
~~~~~~~

::

  -h, --help   help for install

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --config-path string           path to the configuration directory. (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl plugin marketplace <mmctl_plugin_marketplace.rst>`_ 	 - Management of marketplace plugins

