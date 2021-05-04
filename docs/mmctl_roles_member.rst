.. _mmctl_roles_member:

mmctl roles member
------------------

Remove system admin privileges

Synopsis
~~~~~~~~


Remove system admin privileges from some users.

::

  mmctl roles member [users] [flags]

Examples
~~~~~~~~

::

    # You can remove admin privileges from one user
    $ mmctl roles member john_doe

    # Or demote multiple users at the same time
    $ mmctl roles member john_doe jane_doe

Options
~~~~~~~

::

  -h, --help   help for member

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

