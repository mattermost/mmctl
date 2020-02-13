.. _mmctl_channel_search:

mmctl channel search
--------------------

Search a channel

Synopsis
~~~~~~~~


Search a channel by channel name.
Channel can be specified by team. ie. --team myTeam myChannel or by team ID.

::

  mmctl channel search [channel]
  mmctl search --team [team] [channel] [flags]

Examples
~~~~~~~~

::

    channel search myChannel
    channel search --team myTeam myChannel

Options
~~~~~~~

::

  -h, --help          help for search
      --team string   Team name or ID

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

