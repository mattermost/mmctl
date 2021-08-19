.. _mmctl_import_list_incomplete:

mmctl import list incomplete
----------------------------

List incomplete import files uploads

Synopsis
~~~~~~~~


List incomplete import files uploads

::

  mmctl import list incomplete [flags]

Examples
~~~~~~~~

::

    import list incomplete

Options
~~~~~~~

::

  -h, --help   help for incomplete

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --json                         the output format will be in json format
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl import list <mmctl_import_list.rst>`_ 	 - List import files

