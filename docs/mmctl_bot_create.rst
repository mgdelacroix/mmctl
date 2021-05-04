.. _mmctl_bot_create:

mmctl bot create
----------------

Create bot

Synopsis
~~~~~~~~


Create bot.

::

  mmctl bot create [username] [flags]

Examples
~~~~~~~~

::

    bot create testbot

Options
~~~~~~~

::

      --description string    Optional. The description text for the new bot.
      --display-name string   Optional. The display name for the new bot.
  -h, --help                  help for create
      --with-token            Optional. Auto genreate access token for the bot.

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --config-path string           path to the configuration directory. If "$HOME/.mmctl" exists it will take precedence over the default value (default "$XDG_CONFIG_HOME")
      --format string                the format of the command output [plain, json] (default "plain")
      --insecure-sha1-intermediate   allows to use insecure TLS protocols, such as SHA-1
      --insecure-tls-version         allows to use TLS versions 1.0 and 1.1
      --local                        allows communicating with the server through a unix socket
      --strict                       will only run commands if the mmctl version matches the server one

SEE ALSO
~~~~~~~~

* `mmctl bot <mmctl_bot.rst>`_ 	 - Management of bots

