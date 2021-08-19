.. _mmctl_export_job_show:

mmctl export job show
---------------------

Show export job

Synopsis
~~~~~~~~


Show export job

::

  mmctl export job show [exportJobID] [flags]

Examples
~~~~~~~~

::

    export job show

Options
~~~~~~~

::

  -h, --help   help for show

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --json                         the output format will be in json format
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl export job <mmctl_export_job.rst>`_ 	 - List and show export jobs

