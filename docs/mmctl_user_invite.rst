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

      --format string   the format of the command output [plain, json] (default "plain")

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

