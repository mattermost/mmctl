.. _mmctl_channel_users_add:

mmctl channel users add
-----------------------

Add users to channel

Synopsis
~~~~~~~~


Add some users to channel

::

  mmctl channel users add [channel] [users] [flags]

Examples
~~~~~~~~

::

    channel users add myteam:mychannel user@example.com username

Options
~~~~~~~

::

  -h, --help   help for add

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --config-path string           path to the configuration directory. (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl channel users <mmctl_channel_users.rst>`_ 	 - Management of channel users

