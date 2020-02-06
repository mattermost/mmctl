.. _mmctl_user_list:

mmctl user list
---------------

List users

Synopsis
~~~~~~~~


List all users

::

  mmctl user list [flags]

Examples
~~~~~~~~

::

    user list

Options
~~~~~~~

::

      --all            Fetch all users. --page flag will be ignore if provided
  -h, --help           help for list
      --page int       Page number to fetch for the list of users
      --per-page int   Number of users to be fetched (default 200)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

