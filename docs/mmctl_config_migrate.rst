.. _mmctl_config_migrate:

mmctl config migrate
--------------------

Migrate existing config between backends

Synopsis
~~~~~~~~


Migrate a file-based configuration to (or from) a database-based configuration. Point the Mattermost server at the target configuration to start using it

::

  mmctl config migrate [from_config] [to_config] [flags]

Examples
~~~~~~~~

::

  config migrate path/to/config.json "postgres://mmuser:mostest@localhost:5432/mattermost_test?sslmode=disable&connect_timeout=10"

Options
~~~~~~~

::

  -h, --help   help for migrate

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --json                         the output format will be in json format
      --local                        allows communicating with the server through a unix socket
      --quiet                        prevent mmctl to generate output for the commands
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

