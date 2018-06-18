# Tasks for dayreport-send

## Objective
There are sevaral problems related to this program and the approach in general, some of them also related to security.
The objective is to heavily refactor it, make it more flexible, relplace large parts and ultimately only use the skeleton of the current version.

## Tasks
[ ] remove mail handling
[ ] pull hard coded runtime configuration (strings for paths, usernames etc.) out into a config file
[ ] remove sshpass, replace by regular ssh, do connection config via ssh\_config file (see `man ssh_config`)
[ ] change sz configuration from .conf/TOML to .json, using the kugu datamodel
[ ] translate sourcecode to english (comments, variables, etc.)
[ ] 
