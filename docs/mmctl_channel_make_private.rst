.. _mmctl_channel_make_private:

mmctl channel make_private
--------------------------

Set a channel's type to private

Synopsis
~~~~~~~~


Set the type of a channel from Public to Private.
Channel can be specified by [team]:[channel]. ie. myteam:mychannel or by channel ID.

::

  mmctl channel make_private [channel] [flags]

Examples
~~~~~~~~

::

    channel make_private myteam:mychannel

Options
~~~~~~~

::

  -h, --help   help for make_private

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

* `mmctl channel <mmctl_channel.rst>`_ 	 - Management of channels

