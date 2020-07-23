.. _mmctl_permissions_assign:

mmctl permissions assign
------------------------

Assign users to role (EE Only)

Synopsis
~~~~~~~~


Assign users to a role by username (Only works in Enterprise Edition).

::

  mmctl permissions assign [role_name] [username...] [flags]

Examples
~~~~~~~~

::

    permissions assign system_admin john.doe jane.doe
    permissions assign system_manager john.doe jane.doe
    permissions assign system_user_manager john.doe jane.doe
    permissions assign system_read_only_admin john.doe jane.doe

Options
~~~~~~~

::

  -h, --help   help for assign

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

