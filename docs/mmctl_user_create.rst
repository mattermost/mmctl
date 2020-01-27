.. _mmctl_user_create:

mmctl user create
-----------------

Create a user

Synopsis
~~~~~~~~


Create a user

::

  mmctl user create [flags]

Examples
~~~~~~~~

::

    user create --email user@example.com --username userexample --password Password1

Options
~~~~~~~

::

      --email string       Required. The email address for the new user account.
      --firstname string   Optional. The first name for the new user account.
  -h, --help               help for create
      --lastname string    Optional. The last name for the new user account.
      --locale string      Optional. The locale (ex: en, fr) for the new user account.
      --nickname string    Optional. The nickname for the new user account.
      --password string    Required. The password for the new user account.
      --system_admin       Optional. If supplied, the new user will be a system administrator. Defaults to false.
      --username string    Required. Username for the new user account.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

