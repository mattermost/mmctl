.. _mmctl_user_change-password:

mmctl user change-password
--------------------------

Changes a user's password

Synopsis
~~~~~~~~


Changes the password of a user by a new one provided. If the user is changing their own password, the flag --current must indicate the current password. The flag --hashed can be used to indicate that the new password has been introduced already hashed

::

  mmctl user change-password <user> [flags]

Examples
~~~~~~~~

::

    # if you have system permissions, you can change other user's passwords
    $ mmctl user change-password john_doe --password new-password

    # if you are changing your own password, you need to provide the current one
    $ mmctl user change-password my-username --current current-password --password new-password

    # you can ommit these flags to introduce them interactively
    $ mmctl user change-password my-username
    Are you changing your own password? (YES/NO): YES
    Current password:
    New password:

    # if you have system permissions, you can update the password with the already hashed new
    # password. The hashing method should be the same that the server uses internally
    $ mmctl user change-password john_doe --password HASHED_PASSWORD --hashed

Options
~~~~~~~

::

  -c, --current string    The current password of the user. Use only if changing your own password
      --hashed            The supplied password is already hashed
  -h, --help              help for change-password
  -p, --password string   The new password for the user

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

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

