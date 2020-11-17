.. _mmctl_import_list_jobs:

mmctl import list jobs
----------------------

List import jobs

Synopsis
~~~~~~~~


List import jobs

::

  mmctl import list jobs [importJobID] [flags]

Examples
~~~~~~~~

::

   import list jobs

Options
~~~~~~~

::

  -h, --help        help for jobs
      --limit int   The maximum number of jobs to show. (default 10)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl import list <mmctl_import_list.rst>`_ 	 - List import files

