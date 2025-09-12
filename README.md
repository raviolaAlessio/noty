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
To get the assigned task of a user, with status Not Started, Progress,
To Be Tested or Not Done, in the current sprint and export them to a csv use:
```
noty task -a <assignee_name> -s NS,P,TBT,ND --sprint current --outfile out.csv
```

To get the assigned tasks assigned to a given user, with status Done or Not Done
in the 73rd sprint use:
```
noty task -a <assignee_name> -s NS,P,TBT,ND --sprint 73
```

Other flags are available, run `noty -h` or `noty task -h` for more.
