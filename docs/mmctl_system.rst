.. _mmctl_system:

mmctl system
------------

System management

Synopsis
~~~~~~~~


System management commands for interacting with the server state and configuration.

Options
~~~~~~~

::

  -h, --help   help for system

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

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative
* `mmctl system clearbusy <mmctl_system_clearbusy.rst>`_ 	 - Clears the busy state
* `mmctl system getbusy <mmctl_system_getbusy.rst>`_ 	 - Get the current busy state
* `mmctl system setbusy <mmctl_system_setbusy.rst>`_ 	 - Set the busy state to true
* `mmctl system status <mmctl_system_status.rst>`_ 	 - Prints the status of the server
* `mmctl system version <mmctl_system_version.rst>`_ 	 - Prints the remote server version

