.. _mmctl_permissions_add:

mmctl permissions add
---------------------

Add permissions to a role (EE Only)

Synopsis
~~~~~~~~


Add one or more permissions to an existing role (Only works in Enterprise Edition).

::

  mmctl permissions add <role> <permission...> [flags]

Examples
~~~~~~~~

::

    permissions add system_user list_open_teams
    permissions add system_manager sysconsole_read_user_management_channels

Options
~~~~~~~

::

  -h, --help   help for add

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl permissions <mmctl_permissions.rst>`_ 	 - Management of permissions

