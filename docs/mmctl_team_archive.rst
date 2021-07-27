.. _mmctl_team_archive:

mmctl team archive
------------------

Archive teams

Synopsis
~~~~~~~~


Archive some teams.
Archives a team along with all related information including posts from the database.

::

  mmctl team archive [teams] [flags]

Examples
~~~~~~~~

::

    team archive myteam

Options
~~~~~~~

::

      --confirm   Confirm you really want to archive the team and a DB backup has been performed.
  -h, --help      help for archive

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-file-path string      path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl team <mmctl_team.rst>`_ 	 - Management of teams

