.. _mmctl_channel_add:

mmctl channel add
-----------------

Add users to channel

Synopsis
~~~~~~~~


Add some users to channel

::

  mmctl channel add [channel] [users] [flags]

Examples
~~~~~~~~

::

    channel add myteam:mychannel user@example.com username

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

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

