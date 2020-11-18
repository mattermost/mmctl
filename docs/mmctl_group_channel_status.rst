.. _mmctl_group_channel_status:

mmctl group channel status
--------------------------

Show's the group constrain status for the specified channel

Synopsis
~~~~~~~~


Show's the group constrain status for the specified channel

::

  mmctl group channel status [team]:[channel] [flags]

Examples
~~~~~~~~

::

    group channel status myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for status

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

* `mmctl group channel <mmctl_group_channel.rst>`_ 	 - Management of channel groups

