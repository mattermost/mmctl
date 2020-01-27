.. _mmctl_channel_create:

mmctl channel create
--------------------

Create a channel

Synopsis
~~~~~~~~


Create a channel.

::

  mmctl channel create [flags]

Examples
~~~~~~~~

::

    channel create --team myteam --name mynewchannel --display_name "My New Channel"
    channel create --team myteam --name mynewprivatechannel --display_name "My New Private Channel" --private

Options
~~~~~~~

::

      --display_name string   Channel Display Name
      --header string         Channel header
  -h, --help                  help for create
      --name string           Channel Name
      --private               Create a private channel.
      --purpose string        Channel purpose
      --team string           Team name or ID

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")

SEE ALSO
~~~~~~~~

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

