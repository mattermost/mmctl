.. _mmctl_auth_current:

mmctl auth current
------------------

Show current user credentials

Synopsis
~~~~~~~~


Show the currently stored user credentials

::

  mmctl auth current [flags]

Examples
~~~~~~~~

::

    auth current

Options
~~~~~~~

::

  -h, --help   help for current

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl auth <mmctl_auth.rst>`_ 	 - Manages the credentials of the remote Mattermost instances

