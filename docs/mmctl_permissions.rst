.. _mmctl_permissions:

mmctl permissions
-----------------

Management of permissions

Synopsis
~~~~~~~~


Management of permissions

Options
~~~~~~~

::

  -h, --help   help for permissions

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --quiet                        prevent mmctl to generate output for the commands
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative
* `mmctl permissions add <mmctl_permissions_add.rst>`_ 	 - Add permissions to a role (EE Only)
* `mmctl permissions remove <mmctl_permissions_remove.rst>`_ 	 - Remove permissions from a role (EE Only)
* `mmctl permissions reset <mmctl_permissions_reset.rst>`_ 	 - Reset default permissions for role (EE Only)
* `mmctl permissions role <mmctl_permissions_role.rst>`_ 	 - Management of roles

