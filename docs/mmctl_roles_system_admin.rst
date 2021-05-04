.. _mmctl_roles_system_admin:

mmctl roles system_admin
------------------------

Set a user as system admin

Synopsis
~~~~~~~~


Make some users system admins.

::

  mmctl roles system_admin [users] [flags]

Examples
~~~~~~~~

::

    # You can make one user a sysadmin
    $ mmctl roles system_admin john_doe

    # Or promote multiple users at the same time
    $ mmctl roles system_admin john_doe jane_doe

Options
~~~~~~~

::

  -h, --help   help for system_admin

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

* `mmctl roles <mmctl_roles.rst>`_ 	 - Manage user roles

