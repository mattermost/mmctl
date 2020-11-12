.. _mmctl_team_create:

mmctl team create
-----------------

Create a team

Synopsis
~~~~~~~~


Create a team.

::

  mmctl team create [flags]

Examples
~~~~~~~~

::

    team create --name mynewteam --display_name "My New Team"
    team create --name private --display_name "My New Private Team" --private

Options
~~~~~~~

::

      --display_name string   Team Display Name
      --email string          Administrator Email (anyone with this email is automatically a team admin)
  -h, --help                  help for create
      --name string           Team Name
      --private               Create a private team.

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

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

