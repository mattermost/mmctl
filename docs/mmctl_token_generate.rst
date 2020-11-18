.. _mmctl_token_generate:

mmctl token generate
--------------------

Generate token for a user

Synopsis
~~~~~~~~


Generate token for a user

::

  mmctl token generate [user] [description] [flags]

Examples
~~~~~~~~

::

    generate testuser test-token

Options
~~~~~~~

::

  -h, --help   help for generate

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

* `mmctl token <mmctl_token.rst>`_ 	 - manage users' access tokens

