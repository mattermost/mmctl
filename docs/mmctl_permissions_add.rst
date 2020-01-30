.. _mmctl_permissions_add:

mmctl permissions add
---------------------

Add permissions to a role (EE Only)

Synopsis
~~~~~~~~


Add one or more permissions to an existing role (Only works in Enterprise Edition).

::

  mmctl permissions add [role] [permission...] [flags]

Examples
~~~~~~~~

::

    permissions add system_user list_open_teams

Options
~~~~~~~

::

  -h, --help   help for add

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl permissions <mmctl_permissions.rst>`_ 	 - Management of permissions and roles

