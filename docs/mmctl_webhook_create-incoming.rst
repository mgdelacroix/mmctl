.. _mmctl_webhook_create-incoming:

mmctl webhook create-incoming
-----------------------------

Create incoming webhook

Synopsis
~~~~~~~~


create incoming webhook which allows external posting of messages to specific channel

::

  mmctl webhook create-incoming [flags]

Examples
~~~~~~~~

::

    webhook create-incoming --channel [channelID] --user [userID] --display-name [displayName] --description [webhookDescription] --lock-to-channel --icon [iconURL]

Options
~~~~~~~

::

      --channel string        Channel name or ID of the new webhook
      --description string    Incoming webhook description
      --display-name string   Incoming webhook display name
  -h, --help                  help for create-incoming
      --icon string           Icon URL
      --lock-to-channel       Lock to channel
      --owner string          The username, email, or ID of the owner of the webhook
      --user string           The username, email or ID of the user that the webhook should post as

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

* `mmctl webhook <mmctl_webhook.rst>`_ 	 - Management of webhooks

