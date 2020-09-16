.. _mmctl_config_subpath:

mmctl config subpath
--------------------

Update client asset loading to use the configured subpath

Synopsis
~~~~~~~~


Update the hard-coded production client asset paths to take into account Mattermost running on a subpath. This command needs access to the Mattermost assets directory to be able to rewrite the paths.

::

  mmctl config subpath [flags]

Examples
~~~~~~~~

::

    # you can rewrite the assets to use a subpath
    config subpath --assets-dir /opt/mattermost/client --path /mattermost

    # the subpath can have multiple steps
    config subpath --assets-dir /opt/mattermost/client --path /my/custom/subpath

    # or you can fallback to the root path passing /
    config subpath --assets-dir /opt/mattermost/client --path /

Options
~~~~~~~

::

  -a, --assets-dir string   directory of the Mattermost assets in the local filesystem
  -h, --help                help for subpath
  -p, --path string         path to update the assets with

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

