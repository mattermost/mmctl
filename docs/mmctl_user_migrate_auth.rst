.. _mmctl_user_migrate_auth:

mmctl user migrate_auth
-----------------------

Mass migrate user accounts authentication type

Synopsis
~~~~~~~~


Migrates accounts from one authentication provider to another. For example, you can upgrade your authentication provider from email to ldap.

::

  mmctl user migrate_auth [from_auth] [to_auth] [migration-options] [flags]

Examples
~~~~~~~~

::

  user migrate_auth email saml users.json

Options
~~~~~~~

::

      --auto      Automatically migrate all users. Assumes the usernames and emails are identical between Mattermost and SAML services. (saml only)
      --confirm   Confirm you really want to proceed with auto migration. (saml only)
      --force     Force the migration to occur even if there are duplicates on the LDAP server. Duplicates will not be migrated. (ldap only)
  -h, --help      help for migrate_auth

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

