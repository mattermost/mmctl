.. _mmctl_config_reset:

mmctl config reset
------------------

Reset config setting

Synopsis
~~~~~~~~


Resets the value of a config setting by its name in dot notation or a setting section. Accepts multiple values for array settings.

::

  mmctl config reset [flags]

Examples
~~~~~~~~

::

  config reset SqlSettings.DriverName LogSettings

Options
~~~~~~~

::

      --confirm   Confirm you really want to reset all configuration settings to its default value
  -h, --help      help for reset

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --format string   the format of the command output [plain, json] (default "plain")

SEE ALSO
~~~~~~~~

* `mmctl config <mmctl_config.rst>`_ 	 - Configuration

