.. _mmctl_group_channel_enable:

mmctl group channel enable
--------------------------

Enables group constrains in the specified channel

Synopsis
~~~~~~~~


Enables group constrains in the specified channel

::

  mmctl group channel enable [team]:[channel] [flags]

Examples
~~~~~~~~

::

    group channel enable myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for enable

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-file-path string      path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl group channel <mmctl_group_channel.rst>`_ 	 - Management of channel groups

