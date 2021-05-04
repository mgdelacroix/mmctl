.. _mmctl_user_email:

mmctl user email
----------------

Change email of the user

Synopsis
~~~~~~~~


Change the email address associated with a user.

::

  mmctl user email [user] [new email] [flags]

Examples
~~~~~~~~

::

    user email testuser user@example.com

Options
~~~~~~~

::

  -h, --help   help for email

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

