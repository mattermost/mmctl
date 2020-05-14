.. _mmctl_system_clearbusy:

mmctl system clearbusy
----------------------

Clears the busy state

Synopsis
~~~~~~~~


Clear the busy state, which re-enables non-critical services.

::

  mmctl system clearbusy [flags]

Examples
~~~~~~~~

::

    system clearbusy

Options
~~~~~~~

::

  -h, --help   help for clearbusy

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

