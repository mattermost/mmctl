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

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl post <mmctl_post.rst>`_ 	 - Management of posts

