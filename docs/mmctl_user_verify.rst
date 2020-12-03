.. _mmctl_user_verify:

mmctl user verify
-----------------

Verify email of users

Synopsis
~~~~~~~~


Verify the emails of some users.

::

  mmctl user verify [users] [flags]

Examples
~~~~~~~~

::

    user verify user1

Options
~~~~~~~

::

  -h, --help   help for verify

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

