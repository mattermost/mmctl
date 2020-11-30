.. _mmctl_channel_delete:

mmctl channel delete
--------------------

Delete channels

Synopsis
~~~~~~~~


Permanently delete some channels.
Permanently deletes one or multiple channels along with all related information including posts from the database.

::

  mmctl channel delete [channels] [flags]

Examples
~~~~~~~~

::

    channel delete myteam:mychannel

Options
~~~~~~~

::

      --confirm   Confirm you really want to delete the channel and a DB backup has been performed.
  -h, --help      help for delete

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

