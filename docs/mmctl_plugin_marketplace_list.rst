.. _mmctl_plugin_marketplace_list:

mmctl plugin marketplace list
-----------------------------

List marketplace plugins

Synopsis
~~~~~~~~


Gets all plugins from the marketplace server, merging data from locally installed plugins as well as prepackaged plugins shipped with the server

::

  mmctl plugin marketplace list [flags]

Examples
~~~~~~~~

::

    # You can list all the plugins
    $ mmctl plugin marketplace list --all

    # Pagination options can be used too
    $ mmctl plugin marketplace list --page 2 --per-page 10

    # Filtering will narrow down the search
    $ mmctl plugin marketplace list --filter jit

    # You can only retrieve local plugins
    $ mmctl plugin marketplace list --local-only

Options
~~~~~~~

::

      --all             Fetch all plugins. --page flag will be ignore if provided
      --filter string   Filter plugins by ID, name or description
  -h, --help            help for list
      --local-only      Only retrieve local plugins
      --page int        Page number to fetch for the list of users
      --per-page int    Number of users to be fetched (default 200)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl plugin marketplace <mmctl_plugin_marketplace.rst>`_ 	 - Management of marketplace plugins

