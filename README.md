# Usage
A notion utility for task management.

## Configure
Set the `NOTION_API_KEY` environemntet variable.

Configure `noty` using
```
noty configure
```
this will retrieve data on epics ans users and store them in the config file,
if the epics or the users change this command must be run again. Other
customization can be done.

## Use
To get the assigned task of a user, with status Not Started, Progress or
To Be Tested and export them to a csv use.
```
go run . task -a <assignee_name> -s NS,P,TBT --outfile out.csv
```

Other flags are available.
