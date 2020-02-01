.. _mmctl_completion_zsh:

mmctl completion zsh
--------------------

Generates the zsh autocompletion scripts

Synopsis
~~~~~~~~


To load completion, run

. <(mmctl completion zsh)

To configure your zsh shell to load completions for each session, add the above line to your ~/.zshrc


::

  mmctl completion zsh [flags]

Options
~~~~~~~

::

  -h, --help   help for zsh

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl completion <mmctl_completion.rst>`_ 	 - Generates autocompletion scripts for bash and zsh

