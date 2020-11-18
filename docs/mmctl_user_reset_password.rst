.. _mmctl_user_reset_password:

mmctl user reset_password
-------------------------

Send users an email to reset their password

Synopsis
~~~~~~~~


Send users an email to reset their password

::

  mmctl user reset_password [users] [flags]

Examples
~~~~~~~~

::

    user reset_password user@example.com

Options
~~~~~~~

::

  -h, --help   help for reset_password

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

