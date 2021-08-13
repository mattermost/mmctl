.. _mmctl_post_list:

mmctl post list
---------------

List posts for a channel

Synopsis
~~~~~~~~


List posts for a channel

::

  mmctl post list [flags]

Examples
~~~~~~~~

::

    post list myteam:mychannel
    post list myteam:mychannel --number 20

Options
~~~~~~~

::

  -f, --follow       Output appended data as new messages are posted to the channel
  -h, --help         help for list
  -n, --number int   Number of messages to list (default 20)
  -i, --show-ids     Show posts ids

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

* `mmctl post <mmctl_post.rst>`_ 	 - Management of posts

