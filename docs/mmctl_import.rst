.. _mmctl_import:

mmctl import
------------

Management of imports

Synopsis
~~~~~~~~


Management of imports

Options
~~~~~~~

::

  -h, --help   help for import

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

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative
* `mmctl import job <mmctl_import_job.rst>`_ 	 - List and show import jobs
* `mmctl import list <mmctl_import_list.rst>`_ 	 - List import files
* `mmctl import process <mmctl_import_process.rst>`_ 	 - Start an import job
* `mmctl import upload <mmctl_import_upload.rst>`_ 	 - Upload import files

