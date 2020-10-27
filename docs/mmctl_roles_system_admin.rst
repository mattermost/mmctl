.. _mmctl_roles_system_admin:

mmctl roles system_admin
------------------------

Set a user as system admin

Synopsis
~~~~~~~~


Make some users system admins

::

  mmctl roles system_admin [users] [flags]

Examples
~~~~~~~~

::

    roles system_admin user1

Options
~~~~~~~

::

  -h, --help   help for system_admin

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

