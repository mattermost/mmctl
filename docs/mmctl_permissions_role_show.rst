.. _mmctl_permissions_role_show:

mmctl permissions role show
---------------------------

Show the role information

Synopsis
~~~~~~~~


Show all the information about a role.

::

  mmctl permissions role show <role_name> [flags]

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

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl permissions role <mmctl_permissions_role.rst>`_ 	 - Management of roles

