Ansible Agent
-------------
**ARCHIVED:** This project is unused and may be in an incomplete state.
The code is left here just in case it ends up being helpful for someone
who has a similar need and wants some kind of base to write something
for themselves, but this project is unmaintained and probably doesn't
work properly. The remainder of this README is kept as-is for archival
purposes.

The Ansible Agent is a simple daemon used as a transport for Ansible.

Ansible traditionally operates over the SSH daemon that is installed on
all machines that are remotely configured. In general, the SSH daemon
and ControlMaster are good enough for most workflows. Using the SSH
daemon makes starting to use Ansible much easier.

On the other hand, using SSH and ControlMaster can be unreliable
transports. On certain platforms, like Ubuntu, are unusable because the
SSH daemon will randomly exit with exit status 0 and cause random tasks
to fail ([example](https://github.com/ansible/ansible/issues/9174)).

When you manage an entire platform with Ansible, owning and
configuring the machines, then being agentless doesn't really matter. As
long as the agent is easy to install, it's trivial to install one either
embedded in the launched image or by just using Ansible to download a
binary and start it.

Installation
============
To install the daemon, copy the binary to the machine and start it. See
the configuration section below for customizing the daemon.

To have Ansible connect to the agent, copy the file in
`connection_plugins/agent.py` to the connection plugins folder in your
Ansible repository. Set the following setting in your `ansible.cfg`
file.

    [defaults]
    transport = agent

You can also configure the connection type in the playbook by doing:

    ---
    - hosts: all
      connection: agent
      tasks:
        - name: print a greeting message
          comand: echo "Hello, World!"

Configuration
=============
The server configuration file is in toml. A sample configuration is
provided in `conf/defaults.toml`.
