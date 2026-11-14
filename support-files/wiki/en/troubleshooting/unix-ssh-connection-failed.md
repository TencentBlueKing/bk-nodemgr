# Unix SSH connection failed

Test manually with the task's connection details. Replace the placeholders:

```bash
ssh -v -p <port> <user>@<host-address>
# Private-key authentication
ssh -v -i <private-key-file> -p <port> <user>@<host-address>
# Test port connectivity
telnet <host-address> <port>
```

If the manual connection also fails, correct the connection details or resolve network and authentication issues before retrying the task.
