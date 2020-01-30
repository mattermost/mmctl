.. _mmctl_channel_archive:

mmctl channel archive
---------------------

Archive channels

Synopsis
~~~~~~~~


Archive some channels.
Archive a channel along with all related information including posts from the database.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

::

  mmctl channel archive [channels] [flags]

Examples
~~~~~~~~

::

    channel archive myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for archive

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

