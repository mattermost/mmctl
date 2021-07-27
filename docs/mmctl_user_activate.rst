.. _mmctl_user_activate:

mmctl user activate
-------------------

Activate users

Synopsis
~~~~~~~~


Activate users that have been deactivated.

::

  mmctl user activate [emails, usernames, userIds] [flags]

Examples
~~~~~~~~

::

    user activate user@example.com
    user activate username

Options
~~~~~~~

::

  -h, --help   help for activate

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-file-path string      path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

