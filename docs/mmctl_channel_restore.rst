.. _mmctl_channel_restore:

mmctl channel restore
---------------------

Restore some channels, alias: unarchive

Synopsis
~~~~~~~~


Restore a previously deleted channel
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

::

  mmctl channel restore [channels] [flags]

Examples
~~~~~~~~

::

    channel restore|unarchive myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for restore

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

