# Translator

`Translator` is a service bridge for the integration between Docker and external NAT box.

`Translator` automatically configures and unconfigures NAT setting on external NAT 
box for any Docker container by inspecting Docker events(i.e. `macvlan` network 
creation/deletion and container start/stop).

## Pluggable NAT adapter

`Translator` supports pluggable NAT adapter, which currently includes:

* [ios-ansible](https://github.com/tetsusat/translator/tree/master/ios_ansible)
* [ios-restconf](https://github.com/tetsusat/translator/tree/master/ios_restconf) 

## Getting Translator

Get the latest release of `Translator` via [Docker Hub](https://hub.docker.com/r/tetsusat/translator/):

```
$ docker pull tetsusat/translator:latest
```

## Using Translator

Typically, running `Translaor` looks like this:

```
$ docker run -d \
    --name=translator \
    --net=host \
    --volume=/var/run/docker.sock:/tmp/docker.sock \
    --env IOS_MGMT_IP="<ios_mgmt_ip>" \
    --env IOS_USER="<ios_user>" \
    --env IOS_PASS="<ios_pass>" \
    --env IOS_ENABLE_PASS="<ios_enable_pass>" \
    --env OUTSIDE_INTERFACE="GigabitEthernetX" \
    --env INSIDE_INTERFACE="GigabitEthernetY" \
    tetsusat/translator:latest \
      <adapter-name>
```

This is the example for Docker `macvlan` network creation.

```
$ docker network create -d macvlan --subnet=192.168.1.0/24 --gateway=192.168.1.254 -o parent=eth1.10 mytenant
```

This is the example for Docker container running. `ip` option is for NAT Local IP and `global_ip` option is for NAT Global IP.
    
```
$ docker run --net=mytenant --ip=192.168.1.1 -it -l global_ip=10.0.1.1 alpine /bin/sh
```

## Prerequisite setting

NAT device needs some prerequisite setting depending on each NAT device type and each adapter type.

For more information on each, please look at `README` file on each adapter's directory.

## License

MIT
