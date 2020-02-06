.. _mmctl_permissions_remove:

mmctl permissions remove
------------------------

Remove permissions from a role (EE Only)

Synopsis
~~~~~~~~


Remove one or more permissions from an existing role (Only works in Enterprise Edition).

::

  mmctl permissions remove [role] [permission...] [flags]

Examples
~~~~~~~~

::

    permissions remove system_user list_open_teams

Options
~~~~~~~

::

  -h, --help   help for remove

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl permissions <mmctl_permissions.rst>`_ 	 - Management of permissions and roles

