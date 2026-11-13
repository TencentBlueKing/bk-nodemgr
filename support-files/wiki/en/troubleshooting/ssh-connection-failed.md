# Troubleshoot SSH connection failures

Use this guide when the **Detect Host Info via SSH** step fails to connect to a host during Agent installation.

## Check the failure details

Open the failed task and check the error in its execution log. Confirm the target host's login address, SSH port, login user and authentication method in Node Manager.

A host's SSH port is not necessarily the default port 22. Use the port configured for that host.

## Connection timed out, refused or unreachable

Ask your administrator to check:

1. The target host is running and its SSH service is listening on the configured address and port.
2. The Node Manager backend can reach that address and port. A successful connection from your own computer does not prove that the backend can connect.
3. Routing, security groups and firewalls allow the connection from the backend to the target host.

Do not disable firewalls or open the SSH port to the public Internet as a troubleshooting shortcut.

## Authentication failed

Check that:

- The login user exists, is allowed to log in over SSH and is not locked or expired.
- The selected authentication method matches the host's SSH server configuration.
- For password authentication, the credential configured in Node Manager is current.
- For private-key authentication, the configured private key matches a public key authorized for the login user, and the host accepts that key.

Update incorrect credentials in Node Manager, then retry the failed task. Never include passwords or private keys in task comments, screenshots or support requests.

## If the connection still fails

Ask your administrator to check the target host's SSH authentication logs and compare their timestamps with the failed task. When requesting support, provide the task identifier, failure time, error message and authentication method, with credentials and other sensitive information removed.
