.. _mmctl_user_name:

mmctl user name
----------------

Change name of the user

Synopsis
~~~~~~~~


Change name of the user.

::

  mmctl user name [user] [new name] [flags]

Examples
~~~~~~~~

::

    user name testuser newname

Options
~~~~~~~

::

  -h, --help   help for name

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

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

