.. _mmctl_channel_users_add:

mmctl channel users add
-----------------------

Add users to channel

Synopsis
~~~~~~~~


Add some users to channel

::

  mmctl channel users add [channel] [users] [flags]

Examples
~~~~~~~~

::

    channel users add myteam:mychannel user@example.com username

Options
~~~~~~~

::

  -h, --help   help for add

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl channel users <mmctl_channel_users.rst>`_ 	 - Management of channel users

