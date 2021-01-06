.. _mmctl_export_delete:

mmctl export delete
-------------------

Delete export file

Synopsis
~~~~~~~~


Delete export file

::

  mmctl export delete [exportname] [flags]

Examples
~~~~~~~~

::

   export delete export_file.zip

Options
~~~~~~~

::

  -h, --help   help for delete

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl export <mmctl_export.rst>`_ 	 - Management of exports

