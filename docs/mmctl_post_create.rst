.. _mmctl_post_create:

mmctl post create
-----------------

Create a post

Synopsis
~~~~~~~~


Create a post

::

  mmctl post create [flags]

Examples
~~~~~~~~

::

    post create myteam:mychannel --message "some text for the post"

Options
~~~~~~~

::

  -h, --help              help for create
  -m, --message string    Message for the post
  -r, --reply-to string   Post id to reply to

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl post <mmctl_post.rst>`_ 	 - Management of posts

