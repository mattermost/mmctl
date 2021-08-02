.. _mmctl_import_job_list:

mmctl import job list
---------------------

List import jobs

Synopsis
~~~~~~~~


List import jobs

::

  mmctl import job list [flags]

Examples
~~~~~~~~

::

    import job list

Options
~~~~~~~

::

      --all            Fetch all import jobs. --page flag will be ignore if provided
  -h, --help           help for list
      --page int       Page number to fetch for the list of import jobs
      --per-page int   Number of import jobs to be fetched (default 200)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl import job <mmctl_import_job.rst>`_ 	 - List and show import jobs

