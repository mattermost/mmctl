.. _mmctl_auth_delete:

mmctl auth delete
-----------------

Delete an credentials

Synopsis
~~~~~~~~


Delete an credentials by its name

::

  mmctl auth delete [server name] [flags]

Examples
~~~~~~~~

::

    auth delete local-server

Options
~~~~~~~

::

  -h, --help   help for delete

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

* `mmctl auth <mmctl_auth.rst>`_ 	 - Manages the credentials of the remote Mattermost instances

