.. _mmctl_completion_bash:

mmctl completion bash
---------------------

Generates the bash autocompletion scripts

Synopsis
~~~~~~~~


To load completion, run

. <(mmctl completion bash)

To configure your bash shell to load completions for each session, add the above line to your ~/.bashrc


::

  mmctl completion bash [flags]

Options
~~~~~~~

::

  -h, --help   help for bash

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")
      --strict          will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl completion <mmctl_completion.rst>`_ 	 - Generates autocompletion scripts for bash and zsh

