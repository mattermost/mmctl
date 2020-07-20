.. _mmctl_user_convert:

mmctl user convert
------------------

Convert users to bots, or a bot to a user

Synopsis
~~~~~~~~


Convert users to bots, or a bot to a user

::

  mmctl user convert [emails, usernames, userIds] --bot [flags]

Examples
~~~~~~~~

::

    user convert user@example.com anotherUser --bot
  	user convert botusername --email new.email@email.com --password password --user

Options
~~~~~~~

::

      --bot                If supplied, convert users to bots.
      --email string       The email address for the converted user account. Ignored when "user" flag is missing.
      --firstname string   The first name for the converted user account. Ignored when "user" flag is missing.
  -h, --help               help for convert
      --lastname string    The last name for the converted user account. Ignored when "user" flag is missing.
      --locale string      The locale (ex: en, fr) for converted new user account. Ignored when "user" flag is missing.
      --nickname string    The nickname for the converted user account. Ignored when "user" flag is missing.
      --password string    The password for converted new user account. Required when "user" flag is set.
      --system_admin       If supplied, the converted user will be a system administrator. Defaults to false. Ignored when "user" flag is missing.
      --user               If supplied, convert a bot to a user.
      --username string    Username for the converted user account. Ignored when "user" flag is missing.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

