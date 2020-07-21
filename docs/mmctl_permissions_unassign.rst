.. _mmctl_permissions_unassign:

mmctl permissions unassign
--------------------------

Unassign users from role (EE Only)

Synopsis
~~~~~~~~


Unassign users from a role by username (Only works in Enterprise Edition).

::

  mmctl permissions unassign [role_name] [username...] [flags]

Examples
~~~~~~~~

::

    permissions unassign read_only_admin john.doe jane.doe

Options
~~~~~~~

::

  -h, --help   help for unassign

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

