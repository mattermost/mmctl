.. _mmctl_channel_users_remove:

mmctl channel users remove
--------------------------

Remove users from channel

Synopsis
~~~~~~~~


Remove some users from channel

::

  mmctl channel users remove [channel] [users] [flags]

Examples
~~~~~~~~

::

    channel users remove myteam:mychannel user@example.com username
    channel users remove myteam:mychannel --all-users

Options
~~~~~~~

::

      --all-users   Remove all users from the indicated channel.
  -h, --help        help for remove

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl channel users <mmctl_channel_users.rst>`_ 	 - Management of channel users

