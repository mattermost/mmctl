.. _mmctl_permissions_role_unassign:

mmctl permissions role unassign
-------------------------------

Unassign users from role (EE Only)

Synopsis
~~~~~~~~


Unassign users from a role by username (Only works in Enterprise Edition).

::

  mmctl permissions role unassign <role_name> <username...> [flags]

Examples
~~~~~~~~

::

    # Unassign users with usernames 'john.doe' and 'jane.doe' from the role named 'system_admin'.
    permissions unassign system_admin john.doe jane.doe

    # Examples using other system roles
    permissions unassign system_manager john.doe jane.doe
    permissions unassign system_user_manager john.doe jane.doe
    permissions unassign system_read_only_admin john.doe jane.doe

Options
~~~~~~~

::

  -h, --help   help for unassign

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --config-path string           path to the configuration directory. (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl permissions role <mmctl_permissions_role.rst>`_ 	 - Management of roles

