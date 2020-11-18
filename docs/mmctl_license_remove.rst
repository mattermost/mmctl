.. _mmctl_license_remove:

mmctl license remove
--------------------

Remove the current license.

Synopsis
~~~~~~~~


Remove the current license and leave mattermost in Team Edition.

::

  mmctl license remove [flags]

Examples
~~~~~~~~

::

    license remove

Options
~~~~~~~

::

  -h, --help   help for remove

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl license <mmctl_license.rst>`_ 	 - Licensing commands

