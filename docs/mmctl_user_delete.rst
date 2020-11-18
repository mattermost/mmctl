.. _mmctl_user_delete:

mmctl user delete
-----------------

Delete users

Synopsis
~~~~~~~~


Permanently delete some users.
Permanently deletes one or multiple users along with all related information including posts from the database.

::

  mmctl user delete [users] [flags]

Examples
~~~~~~~~

::

    user delete user@example.com

Options
~~~~~~~

::

      --confirm   Confirm you really want to delete the user and a DB backup has been performed
  -h, --help      help for delete

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

