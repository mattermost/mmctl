.. _mmctl_roles_member:

mmctl roles member
------------------

Remove system admin privileges

Synopsis
~~~~~~~~


Remove system admin privileges from some users.

::

  mmctl roles member [users] [flags]

Examples
~~~~~~~~

::

    roles member user1

Options
~~~~~~~

::

  -h, --help   help for member

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl roles <mmctl_roles.rst>`_ 	 - Management of user roles

