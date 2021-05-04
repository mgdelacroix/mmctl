.. _mmctl_user_demote:

mmctl user demote
-----------------

Demote users to guests

Synopsis
~~~~~~~~


Convert a regular user into a guest.

::

  mmctl user demote [users] [flags]

Examples
~~~~~~~~

::

    user demote user1 user2

Options
~~~~~~~

::

  -h, --help   help for demote

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

