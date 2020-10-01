.. _mmctl_logs:

mmctl logs
----------

Display logs in a human-readable format

Synopsis
~~~~~~~~


Display logs in a human-readable format. The "--format" flag cannot be used here.

::

  mmctl logs [flags]

Options
~~~~~~~

::

  -h, --help         help for logs
  -l, --logrus       Use logrus for formatting.
  -n, --number int   Number of log lines to retrieve. (default 200)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative

