.. _mmctl_import_process:

mmctl import process
--------------------

Start an import job

Synopsis
~~~~~~~~


Start an import job

::

  mmctl import process [importname] [flags]

Examples
~~~~~~~~

::

    import process 35uy6cwrqfnhdx3genrhqqznxc_import.zip

Options
~~~~~~~

::

  -h, --help   help for process

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

* `mmctl import <mmctl_import.rst>`_ 	 - Management of imports

