.. _mmctl_auth_set:

mmctl auth set
--------------

Set the credentials to use

Synopsis
~~~~~~~~


Set an credentials to use in the following commands

::

  mmctl auth set [server name] [flags]

Examples
~~~~~~~~

::

    auth set local-server

Options
~~~~~~~

::

  -h, --help   help for set

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-file-path string      path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl auth <mmctl_auth.rst>`_ 	 - Manages the credentials of the remote Mattermost instances

