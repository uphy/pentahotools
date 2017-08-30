# Pentaho Tools

[![GitHub license](https://img.shields.io/badge/license-Apache%202-blue.svg)](https://raw.githubusercontent.com/uphy/pentahotools/master/LICENSE)
[![wercker status](https://app.wercker.com/status/b411f16ffe5211a9c3578fb1cf9322e2/s/master "wercker status")](https://app.wercker.com/project/byKey/b411f16ffe5211a9c3578fb1cf9322e2)

A Pentaho API CLI client app written in Go.

# Usage

Get the 'pentahotools' binary from [Releases page](https://github.com/uphy/pentahotools/releases).

'pentahotools' consists of subcommands.

Subcommands:
|Subcommand |Description|
|:------------------------|:----------|
|[carte](#carte)          |Manage the jobs/transformations of the DI(Carte) server.|
|[datasource](#datasource)|Manage the datasources of BA/DI server.|
|[file](#file)            |Manage the repository files of BA/DI server.|
|[userrole](#userrole)    |Manage the users and roles of BA/DI server.|

## Global flags

In addition to the flags for subcommands, there's some global flags.

Subcommands:
|Flag |Description|
|:------------------------|:----------|
|-l|The URL of the Pentaho server. (e.g., http://localhost:8080/pentaho)|
|-u|Login user name.|
|-p|Login password.|

## Subcommands

### [carte](#carte)

### [datasource](#datasource)

### [file](#file)

### [userrole](#userrole)

## Pentaho Tools Shell

You can enter to the pentaho tools shell by executing the command without arguments.

```bash
$ pentahotools -l http://localhost:8080/pentaho -u admin -p password
Entering multiple command mode.
Input 'exit' to exit this command.
>
```

# References

https://github.com/pentaho/pentaho-platform/blob/7.1.0.1/extensions/src/main/java/org/pentaho/platform/web/http/api/resources/

https://github.com/pentaho/pentaho-kettle/tree/7.1.0.1/engine/src/org/pentaho/di/www
