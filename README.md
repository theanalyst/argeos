# Argeos

Argeos is a daemon to trigger diagnostic information for stuck processes, it
uses a plugin architecture where plugins support diagnostic dump & healthcheck
commands. Currently we have a shell and a network plugin.

Running the command:
```
argeos -c config.json -logfile=/var/log/argeos/argeos.log
```
