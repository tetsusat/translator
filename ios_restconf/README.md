# Translator adapter for IOS RESTCONF

This adapter is for `Cisco IOS RESTCONF`.

## Using Translator with ios-restconf adapter

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
      ios-restconf
```

## Prerequisite setting for ios-restconf adapter

This is a prerequisite NAT setting of Cisco IOS.

```
interface GigabitEthernetX  # NAT outside interface
 ip nat outside
!
ip access-list extended NAT
 permit ip any any
!
route-map NAT permit 10
 set global
```

Cisco IOS RESTCONF needs the setting below too.

```
username <username> privilege 15 password 0 <password>
!
restconf
```
