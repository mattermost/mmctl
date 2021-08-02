.. _mmctl_license_upload:

mmctl license upload
--------------------

Upload a license.

Synopsis
~~~~~~~~


Upload a license. Replaces current license.

::

  mmctl license upload [license] [flags]

Examples
~~~~~~~~

::

    license upload /path/to/license/mylicensefile.mattermost-license

Options
~~~~~~~

::

  -h, --help   help for upload

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

* `mmctl license <mmctl_license.rst>`_ 	 - Licensing commands

