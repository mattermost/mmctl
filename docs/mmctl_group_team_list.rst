.. _mmctl_group_team_list:

mmctl group team list
---------------------

List team groups

Synopsis
~~~~~~~~


List the groups associated with a team

::

  mmctl group team list [team] [flags]

Examples
~~~~~~~~

::

    group team list myteam

Options
~~~~~~~

::

  -h, --help   help for list

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl group team <mmctl_group_team.rst>`_ 	 - Management of team groups

