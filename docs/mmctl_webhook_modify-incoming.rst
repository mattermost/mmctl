.. _mmctl_webhook_modify-incoming:

mmctl webhook modify-incoming
-----------------------------

Modify incoming webhook

Synopsis
~~~~~~~~


Modify existing incoming webhook by changing its title, description, channel or icon url

::

  mmctl webhook modify-incoming [flags]

Examples
~~~~~~~~

::

    webhook modify-incoming [webhookID] --channel [channelID] --display-name [displayName] --description [webhookDescription] --lock-to-channel --icon [iconURL]

Options
~~~~~~~

::

      --channel string        Channel ID
      --description string    Incoming webhook description
      --display-name string   Incoming webhook display name
  -h, --help                  help for modify-incoming
      --icon string           Icon URL
      --lock-to-channel       Lock to channel

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

