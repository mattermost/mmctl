.. _mmctl_team_delete:

mmctl team delete
-----------------

Delete teams

Synopsis
~~~~~~~~


Permanently delete some teams.
Permanently deletes a team along with all related information including posts from the database.

::

  mmctl team delete [teams] [flags]

Examples
~~~~~~~~

::

    team delete myteam

Options
~~~~~~~

::

      --confirm   Confirm you really want to delete the team and a DB backup has been performed.
  -h, --help      help for delete

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

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

