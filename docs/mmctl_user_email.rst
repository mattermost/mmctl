.. _mmctl_user_email:

mmctl user email
----------------

Change email of the user

Synopsis
~~~~~~~~


Change email of the user.

::

  mmctl user email [user] [new email] [flags]

Examples
~~~~~~~~

::

    user email testuser user@example.com

Options
~~~~~~~

::

  -h, --help   help for email

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

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

