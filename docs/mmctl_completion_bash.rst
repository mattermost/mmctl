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

      --config-path string           path to search for '.mmctl' configuration file (default "$HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl completion <mmctl_completion.rst>`_ 	 - Generates autocompletion scripts for bash and zsh

