.. _mmctl_group_list-ldap:

mmctl group list-ldap
---------------------

List LDAP groups

Synopsis
~~~~~~~~


List LDAP groups

::

  mmctl group list-ldap [flags]

Examples
~~~~~~~~

::

    group list-ldap

Options
~~~~~~~

::

  -h, --help   help for list-ldap

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --disable-pager                disables paged output
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl group <mmctl_group.rst>`_ 	 - Management of groups

