.. _mmctl_user_convert:

mmctl user convert
------------------

Convert users to bots, or a bot to a user

Synopsis
~~~~~~~~


Convert user accounts to bots or convert bots to user accounts.

::

  mmctl user convert (--bot [emails] [usernames] [userIds] | --user <username> --password PASSWORD [--email EMAIL]) [flags]

Examples
~~~~~~~~

::

    # you can convert a user to a bot providing its email, id or username
    $ mmctl user convert user@example.com --bot

    # or multiple users in one go
    $ mmctl user convert user@example.com anotherUser --bot

    # you can convert a bot to a user specifying the email and password that the user will have after conversion
    $ mmctl user convert botusername --email new.email@email.com --password password --user

Options
~~~~~~~

::

      --bot                If supplied, convert users to bots
      --email string       The email address for the converted user account. Required when the "bot" flag is set
      --firstname string   The first name for the converted user account. Required when the "bot" flag is set
  -h, --help               help for convert
      --lastname string    The last name for the converted user account. Required when the "bot" flag is set
      --locale string      The locale (ex: en, fr) for converted new user account. Required when the "bot" flag is set
      --nickname string    The nickname for the converted user account. Required when the "bot" flag is set
      --password string    The password for converted new user account. Required when "user" flag is set
      --system_admin       If supplied, the converted user will be a system administrator. Defaults to false. Required when the "bot" flag is set
      --user               If supplied, convert a bot to a user
      --username string    Username for the converted user account. Required when the "bot" flag is set

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

* `mmctl user <mmctl_user.rst>`_ 	 - Management of users

