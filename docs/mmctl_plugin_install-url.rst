.. _mmctl_plugin_install-url:

mmctl plugin install-url
------------------------

Install plugin from url

Synopsis
~~~~~~~~


Supply one or multiple URLs to plugins compressed in a .tar.gz file. Plugins must be enabled in the server's config settings

::

  mmctl plugin install-url <url>... [flags]

Examples
~~~~~~~~

::

    # You can install one plugin
    $ mmctl plugin install-url https://example.com/mattermost-plugin.tar.gz

    # Or install multiple in one go
    $ mmctl plugin install-url https://example.com/mattermost-plugin-one.tar.gz https://example.com/mattermost-plugin-two.tar.gz

Options
~~~~~~~

::

  -f, --force   overwrite a previously installed plugin with the same ID, if any
  -h, --help    help for install-url

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

* `mmctl plugin <mmctl_plugin.rst>`_ 	 - Management of plugins

