.. _mmctl_user_deactivate:

mmctl user deactivate
---------------------

Deactivate users

Synopsis
~~~~~~~~


Deactivate users. Deactivated users are immediately logged out of all sessions and are unable to log back in.

::

  mmctl user deactivate [emails, usernames, userIds] [flags]

Examples
~~~~~~~~

::

    user deactivate user@example.com
    user deactivate username

Options
~~~~~~~

::

  -h, --help   help for deactivate

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

