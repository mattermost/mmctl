.. _mmctl_config_edit:

mmctl config edit
-----------------

Edit the config

Synopsis
~~~~~~~~


Opens the editor defined in the EDITOR environment variable to modify the server's configuration and then uploads it

::

  mmctl config edit [flags]

Examples
~~~~~~~~

::

  config edit

Options
~~~~~~~

::

  -h, --help   help for edit

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config string                path to search for '.mmctl' configuration file (default "$HOME/.config/mmctl")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

