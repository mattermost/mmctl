.. _mmctl_webhook_show:

mmctl webhook show
------------------

Show a webhook

Synopsis
~~~~~~~~


Show the webhook specified by [webhookId]

::

  mmctl webhook show [webhookId] [flags]

Examples
~~~~~~~~

::

    webhook show w16zb5tu3n1zkqo18goqry1je

Options
~~~~~~~

::

  -h, --help   help for show

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl webhook <mmctl_webhook.rst>`_ 	 - Management of webhooks

