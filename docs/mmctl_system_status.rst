.. _mmctl_system_status:

mmctl system status
-------------------

Prints the status of the server

Synopsis
~~~~~~~~


Prints the server status calculated using several basic server healthchecks

::

  mmctl system status [flags]

Examples
~~~~~~~~

::

    system status

Options
~~~~~~~

::

  -h, --help   help for status

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl system <mmctl_system.rst>`_ 	 - System management

