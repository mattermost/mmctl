.. _mmctl_token_revoke:

mmctl token revoke
------------------

Revoke tokens for a user

Synopsis
~~~~~~~~


Revoke tokens for a user

::

  mmctl token revoke [token-ids] [flags]

Examples
~~~~~~~~

::

    revoke testuser test-token-id

Options
~~~~~~~

::

  -h, --help   help for revoke

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

* `mmctl token <mmctl_token.rst>`_ 	 - manage users' access tokens

