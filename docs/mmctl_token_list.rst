.. _mmctl_token_list:

mmctl token list
----------------

List users tokens

Synopsis
~~~~~~~~


List the tokens of a user

::

  mmctl token list [user] [flags]

Examples
~~~~~~~~

::

    user tokens testuser

Options
~~~~~~~

::

      --active         List only active tokens (default true)
      --all            Fetch all tokens. --page flag will be ignore if provided
  -h, --help           help for list
      --inactive       List only inactive tokens
      --page int       Page number to fetch for the list of users
      --per-page int   Number of users to be fetched (default 200)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl token <mmctl_token.rst>`_ 	 - manage users' access tokens

