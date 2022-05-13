# Lagoon - Simple Linux package repository mirror

[![Go Report Card](https://goreportcard.com/badge/github.com/klaasjand/lagoon)](https://goreportcard.com/report/github.com/klaasjand/lagoon)

A lagoon is a shallow stretch of water separated from the sea by a reef or 
sandbank. Lagoon can be used to mirror package repositories or parts of these 
repositories from the internet hence the name Lagoon.

When running Linux servers in an enterprise environment it is useful to have 
all servers running the same versions of software. In order to accomplish this, 
package repositories must be frozen at a certain point in time because 
normal (public) repositories are constantly updated and 'on the move'.

Lagoon can be used to set up a local mirror for upstream OS package repositories 
and makes it possible to capture certain points in time or so called snapshots. 
It also provides the latest snapshot or version of the upstream repository. 
Lagoon is only capable of providing very basic functionality and is not a 
replacement for Red Hat's 
[Satellite](https://www.redhat.com/en/technologies/management/satellite), 
Foreman's [Katello](https://theforeman.org/plugins/katello/) or 
[Pulp](https://pulpproject.org/).

Lagoon synchronizes with the remote repository and stores the files in 
the `upstream` folder. When synchronization is complete a point in time 
snapshot is made in the `staging` folder according to the following pattern 
`20060102` (therefore the snapshot resolution cannot be smaller than a day). 
After the point in time snapshot is created, it is published in the `public` 
folder using the same pattern. The last snapshot is also published as `latest`.

## Running Lagoon

Lagoon can be run as a standalone golang binary or from a Docker container. 
A Docker Compose [file](deployments/docker/docker-compose.yml) is included for 
reference.

Run Lagoon from a Docker container:
```shell
docker run --name lagoon -v $PWD/lagoon.yml:/etc/lagoon/lagoon.yml lagoon:latest
```

### Requirements

The following dependencies are needed for running Lagoon:
* `rsync`
* `yum-utils`
* `createrepo`

### Supported synchronisation methods

| Sync method  | Supported | Status |
|-|-|-|
| Rsync | yes | |
| RPM reposync | beta | Basic sync. TODO: implement errata support |

### File storage

The treeview below shows how snapshots are stored, for example `repo1` consists 
of daily snapshots and `repo2` consists of weekly snapshots. Each folder 
contains the files which were downloaded from the upstream repository at that 
certain moment in time. Hardlinks are being used for efficient storage and to 
make sure files do not disappear from a staging snapshot when they are deleted 
from the upstream content. Staging snapshots are published with symlinks, the 
latest snapshot always points to the last staging snapshot.
```
/var/lib/lagoon/
|-- public
|   |-- repo1
|   |   |-- 20220126 -> /var/lib/lagoon/staging/repo1/20220126
|   |   |-- 20220127 -> /var/lib/lagoon/staging/repo1/20220127
|   |   |-- 20220128 -> /var/lib/lagoon/staging/repo1/20220128
|   |   |-- 20220129 -> /var/lib/lagoon/staging/repo1/20220129
|   |   |-- 20220130 -> /var/lib/lagoon/staging/repo1/20220130
|   |   `-- latest -> /var/lib/lagoon/staging/repo1/20220130
|   `-- repo2
|       |-- 20220115 -> /var/lib/lagoon/staging/repo2/20220115
|       |-- 20220122 -> /var/lib/lagoon/staging/repo2/20220122
|       |-- 20220129 -> /var/lib/lagoon/staging/repo2/20220129
|       `-- latest -> /var/lib/lagoon/staging/repo2/20220129
|-- staging
|   |-- repo1
|   |   |-- 20220126
|   |   |-- 20220127
|   |   |-- 20220128
|   |   |-- 20220129
|   |   `-- 20220130
|   `-- repo2
|       |-- 20220115
|       |-- 20220122
|       `-- 20220129
`-- upstream
    |-- repo1
    |   |-- file_1
    |   `-- file_n
    `-- repo2
        |-- file_1
        `-- file_n
```
Lagoon can also take care of automatically freeing up diskspace by removing 
snapshots which aren't used anymore. This can be configured by telling Lagoon 
how much snapshots it has to keep for a certain repository.

### Configuration

See [lagoon.example.yml](lagoon.example.yml) for example configuration.

```yaml
repositories:
  - id: docker-ce_centos-7 # Unique id
    # Name of the repo
    name: Docker CE CentOS 7 x86_64
    # Type of remote repository (rsync or reposync)
    type: reposync
    # Upstream rsync url or reposync multiline string with yum repo config
    src: |
      [docker-ce-stable-centos7]
      baseurl = https://download.docker.com/linux/centos/7/x86_64/stable
      enabled = 1
      gpgcheck = 1
      gpgkey = https://download.docker.com/linux/centos/gpg
      name = Docker CE Stable - x86_64
    # Destination the repo (absolute path)
    dest: /var/lib/lagoon
    # Cron sync expression see: https://github.com/robfig/cron
    cron: "*/10 * * * * *"
    # Number of snapshots to keep
    snapshots: 52
    # List of directories to exclude from rsync
    #exclude: []
```

### Logging and monitoring

By default Lagoon logs to stdout using JSON format. In order to enable debug 
logging start Lagoon with `-d` parameter. For human readable logging start 
Lagoon with `-h` parameter. Each separate sync job can be identified by the 
repository name and a unique ID.

Lagoon can be monitored with Prometheus and exposes its metrics on port `9000` 
at `/metrics`. In addition to the standard golang metrics the following Lagoon 
specific metrics are exposed:

| Metric                       | Description                    |
|------------------------------|--------------------------------|
| lagoon_sync_total            | The total number of repo syncs |
| lagoon_sync_duration_seconds | The sync duration              |

## Building Lagoon

We will use Docker as a build environment so local installation of build tools 
is not required. Execute the commands from the root of the project.

Building a Docker image:
```shell
docker build -f build/Dockerfile -t lagoon .
```
You can either run a Docker container from the image or use the following 
command to extract the `lagoon` binary to run it separately:
```shell
docker run --rm -v $PWD:/build lagoon cp /etc/lagoon/lagoon /build
```

## Developing Lagoon

The preferred method for developing Lagoon is using VSCode with the 'Remote - 
Containers' extension `ms-vscode-remote.remote-containers`. A `.devcontainer` 
context is included with the project in order to open it in a remote container 
and get your development environment up-and-running quickly.
