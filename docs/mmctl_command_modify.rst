.. _mmctl_command_modify:

mmctl command modify
--------------------

Modify a slash command

Synopsis
~~~~~~~~


Modify a slash command. Commands can be specified by command ID.

::

  mmctl command modify [commandID] [flags]

Examples
~~~~~~~~

::

    command modify commandID --title MyModifiedCommand --description "My Modified Command Description" --trigger-word mycommand --url http://localhost:8000/my-slash-handler --creator myusername --response-username my-bot-username --icon http://localhost:8000/my-slash-handler-bot-icon.png --autocomplete --post

Options
~~~~~~~

::

      --autocomplete               Show Command in autocomplete list
      --autocompleteDesc string    Short Command Description for autocomplete list
      --autocompleteHint string    Command Arguments displayed as help in autocomplete list
      --creator string             Command Creator's username, email or id (required)
      --description string         Command Description
  -h, --help                       help for modify
      --icon string                Command Icon URL
      --post                       Use POST method for Callback URL
      --response-username string   Command Response Username
      --title string               Command Title
      --trigger-word string        Command Trigger Word (required)
      --url string                 Command Callback URL (required)

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl command <mmctl_command.rst>`_ 	 - Management of slash commands

