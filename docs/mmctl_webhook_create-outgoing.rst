.. _mmctl_webhook_create-outgoing:

mmctl webhook create-outgoing
-----------------------------

Create outgoing webhook

Synopsis
~~~~~~~~


create outgoing webhook which allows external posting of messages from a specific channel

::

  mmctl webhook create-outgoing [flags]

Examples
~~~~~~~~

::

    webhook create-outgoing --team myteam --user myusername --display-name mywebhook --trigger-word "build" --trigger-word "test" --url http://localhost:8000/my-webhook-handler
  	webhook create-outgoing --team myteam --channel mychannel --user myusername --display-name mywebhook --description "My cool webhook" --trigger-when start --trigger-word build --trigger-word test --icon http://localhost:8000/my-slash-handler-bot-icon.png --url http://localhost:8000/my-webhook-handler --content-type "application/json"

Options
~~~~~~~

::

      --channel string             Channel name or ID
      --content-type string        Content-type
      --description string         Outgoing webhook description
      --display-name string        Outgoing webhook display name
  -h, --help                       help for create-outgoing
      --icon string                Icon URL
      --owner string               The username, email, or ID of the owner of the webhook
      --team string                Team name or ID
      --trigger-when string        When to trigger webhook (exact: for first word matches a trigger word exactly, start: for first word starts with a trigger word) (default "exact")
      --trigger-word stringArray   Word to trigger webhook
      --url stringArray            Callback URL
      --user string                The username, email or ID of the user that the webhook should post as

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

