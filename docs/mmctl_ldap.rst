.. _mmctl_ldap:

mmctl ldap
----------

LDAP related utilities

Synopsis
~~~~~~~~


LDAP related utilities

Options
~~~~~~~

::

  -h, --help   help for ldap

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

* `mmctl <mmctl.rst>`_ 	 - Remote client for the Open Source, self-hosted Slack-alternative
* `mmctl ldap idmigrate <mmctl_ldap_idmigrate.rst>`_ 	 - Migrate LDAP IdAttribute to new value
* `mmctl ldap sync <mmctl_ldap_sync.rst>`_ 	 - Synchronize now

