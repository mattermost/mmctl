.. _mmctl_permissions_show:

mmctl permissions show
----------------------

Show the role information

Synopsis
~~~~~~~~


Show all the information about a role.

::

  mmctl permissions show [role_name] [flags]

Examples
~~~~~~~~

::

    permissions show system_user

Options
~~~~~~~

::

  -h, --help   help for show

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl permissions <mmctl_permissions.rst>`_ 	 - Management of permissions and roles

