.. _mmctl_channel_make_private:

mmctl channel make_private
--------------------------

Set a channel's type to private

Synopsis
~~~~~~~~


Set the type of a channel from public to private.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

::

  mmctl channel make_private [channel] [flags]

Examples
~~~~~~~~

::

    channel make_private myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for make_private

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")

SEE ALSO
~~~~~~~~

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

