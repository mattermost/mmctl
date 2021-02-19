.. _mmctl_integrity:

mmctl integrity
---------------

Check database records integrity.

Synopsis
~~~~~~~~


Perform a relational integrity check which returns information about any orphaned record found.

::

  mmctl integrity [flags]

Options
~~~~~~~

::

      --confirm   Confirm you really want to run a complete integrity check that may temporarily harm system performance
  -h, --help      help for integrity
  -v, --verbose   Show detailed information on integrity check results

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

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative

