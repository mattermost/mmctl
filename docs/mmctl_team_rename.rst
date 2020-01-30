.. _mmctl_team_rename:

mmctl team rename
-----------------

Rename team

Synopsis
~~~~~~~~


Rename an existing team

::

  mmctl team rename [team] [flags]

Examples
~~~~~~~~

::

    team rename old-team --name 'new-team' --display_name 'New Display Name'
    team rename old-team --name 'new-team'
    team rename old-team --display_name 'New Display Name'

Options
~~~~~~~

::

      --display_name string   Team Display Name
  -h, --help                  help for rename
      --name string           Old Team Name

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

