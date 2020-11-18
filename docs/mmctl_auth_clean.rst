.. _mmctl_auth_clean:

mmctl auth clean
----------------

Clean all credentials

Synopsis
~~~~~~~~


Clean the currently stored credentials

::

  mmctl auth clean [flags]

Examples
~~~~~~~~

::

    auth clean

Options
~~~~~~~

::

  -h, --help   help for clean

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

* `mmctl auth <mmctl_auth.rst>`_ 	 - Manages the credentials of the remote Mattermost instances

