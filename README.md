# Keysafe
A simple program for managing secrets in the Linux keyring.

```
$ keysafe help
keysafe: keyctl password management.

keysafe manages authentication data on behalf of other applications.
Applications use the keyctl package also defined here.

Usage:
  keysafe [command]

Available Commands:
  del         Delete entries from the keyring
  dump        Dump all keys as JSON
  get         Get the contents of a named key
  help        Help about any command
  list        List known keys
  reap        Remove expired keys from a keyring
  restore     Restore a keysafe saved by "keysafe dump"
  set         Set values in an entry

Flags:
  -h, --help             help for keysafe
  -k, --keyring string   keyring name (default "keysafe")

Use "keysafe [command] --help" for more information about a command.
```
