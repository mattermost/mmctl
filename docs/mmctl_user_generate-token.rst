.. _mmctl_user_generate-token:

mmctl user generate-token
-------------------------

Generate token for a user

Synopsis
~~~~~~~~


Generate token for a user

::

  mmctl user generate-token [user] [description] [flags]

Examples
~~~~~~~~

::

    generate-token testuser test-token

Options
~~~~~~~

::

  -h, --help   help for generate-token

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

