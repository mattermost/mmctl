.. _mmctl_import_upload:

mmctl import upload
-------------------

Upload import files

Synopsis
~~~~~~~~


Upload import files

::

  mmctl import upload [filepath] [flags]

Examples
~~~~~~~~

::

    import upload import_file.zip

Options
~~~~~~~

::

  -h, --help            help for upload
      --resume          Set to true to resume an incomplete import upload.
      --upload string   The ID of the import upload to resume.

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

* `mmctl import <mmctl_import.rst>`_ 	 - Management of imports

