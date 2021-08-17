.. _mmctl_user_resetmfa:

mmctl user resetmfa
-------------------

Turn off MFA

Synopsis
~~~~~~~~


Turn off multi-factor authentication for a user.
If MFA enforcement is enabled, the user will be forced to re-enable MFA as soon as they log in.

::

  mmctl user resetmfa [users] [flags]

Examples
~~~~~~~~

::

    user resetmfa user@example.com

Options
~~~~~~~

::

  -h, --help   help for resetmfa

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to the configuration file (default "$XDG_CONFIG_HOME/mmctl/config")
      --config-path string           path to the configuration directory. (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one
      --suppress-warnings            disables printing warning messages

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

