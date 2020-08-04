# LIGHTKEEPER

## DESCRIPTION

Lightkeeper is a tool that provides an easy and accesible way of making backups of Docker container data. It also allows to recover from previous backups.

It was first written in Python but now is being developed fully in Go.

## FEATURES

* Backup volumes and binds
* Launch containers with backup data
* Configuration can be specified in a YAML file
* Cron job to backup data from containers automatically.

## FUTURE WORK

* Logging
* Error recovery
* Support for non standard file systems
* Support for deploying stacks 
* Remote copy of backups
* Distributed version to work with several systems
* API version

## USAGE

`go build`

`./lightkeeper ___`


## EXAMPLE CONFIG

```
containers:
    -
        name: my_web_server
        image: nginx
        ports:
            8080: "80/tcp"
        mounts:
            - 
                ID: 0
                type: bind
                from: /home/foo/test
                dstpath: /tmp/test1
            
            -
                ID: 1
                type: volume
                from: test
                dstpath: /tmp/test2
        env:
            - "MYVAR=/etc/bin"
```