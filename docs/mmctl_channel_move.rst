.. _mmctl_channel_move:

mmctl channel move
------------------

Moves channels to the specified team

Synopsis
~~~~~~~~


Moves the provided channels to the specified team.
Validates that all users in the channel belong to the target team. Incoming/Outgoing webhooks are moved along with the channel.
Channels can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

::

  mmctl channel move [team] [channels] [flags]

Examples
~~~~~~~~

::

    channel move newteam oldteam:mychannel

Options
~~~~~~~

::

      --force   Remove users that are not members of target team before moving the channel.
  -h, --help    help for move

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

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

