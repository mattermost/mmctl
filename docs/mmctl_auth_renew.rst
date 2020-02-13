.. _mmctl_auth_renew:

mmctl auth renew
----------------

Renews a set of credentials

Synopsis
~~~~~~~~


Renews the credentials for a given server

::

  mmctl auth renew [flags]

Examples
~~~~~~~~

::

    auth renew local-server

Options
~~~~~~~

::

  -a, --access-token string   Access token to use instead of username/password
  -h, --help                  help for renew
  -m, --mfa-token string      MFA token for the credentials
  -p, --password string       Password for the credentials

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl auth <mmctl_auth.rst>`_ 	 - Manages the credentials of the remote Mattermost instances

