.. _mmctl_group_channel_disable:

mmctl group channel disable
---------------------------

Disables group constrains in the specified channel

Synopsis
~~~~~~~~


Disables group constrains in the specified channel

::

  mmctl group channel disable [team]:[channel] [flags]

Examples
~~~~~~~~

::

    group channel disable myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for disable

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl group channel <mmctl_group_channel.rst>`_ 	 - Management of channel groups

