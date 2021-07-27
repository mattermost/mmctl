.. _mmctl_user_invite:

mmctl user invite
-----------------

Send user an email invite to a team.

Synopsis
~~~~~~~~


Send user an email invite to a team.
You can invite a user to multiple teams by listing them.
You can specify teams by name or ID.

::

  mmctl user invite [email] [teams] [flags]

Examples
~~~~~~~~

::

    user invite user@example.com myteam
    user invite user@example.com myteam1 myteam2

Options
~~~~~~~

::

  -h, --help   help for invite

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

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

