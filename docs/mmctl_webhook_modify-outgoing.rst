.. _mmctl_webhook_modify-outgoing:

mmctl webhook modify-outgoing
-----------------------------

Modify outgoing webhook

Synopsis
~~~~~~~~


Modify existing outgoing webhook by changing its title, description, channel, icon, url, content-type, and triggers

::

  mmctl webhook modify-outgoing [flags]

Examples
~~~~~~~~

::

    webhook modify-outgoing [webhookId] --channel [channelId] --display-name [displayName] --description "New webhook description" --icon http://localhost:8000/my-slash-handler-bot-icon.png --url http://localhost:8000/my-webhook-handler --content-type "application/json" --trigger-word test --trigger-when start

Options
~~~~~~~

::

      --channel string             Channel name or ID
      --content-type string        Content-type
      --description string         Outgoing webhook description
      --display-name string        Outgoing webhook display name
  -h, --help                       help for modify-outgoing
      --icon string                Icon URL
      --trigger-when string        When to trigger webhook (exact: for first word matches a trigger word exactly, start: for first word starts with a trigger word)
      --trigger-word stringArray   Word to trigger webhook
      --url stringArray            Callback URL

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

