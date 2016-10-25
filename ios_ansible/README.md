# Translator adapter for IOS Ansible

This adapter is for `Cisco IOS Ansible`.

## Using Translator with ios-ansible adapter

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
      ios-ansible
```

## Prerequisite setting for ios-ansible adapter

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

Cisco IOS Ansible needs the setting below too.

```
hostname <hostname>
!
enable password <enable_password>
!
username <username> privilege 15 password 0 <password>
!
ip domain-name <domain_name>
!
line vty 0 4
 login local
 transport input ssh
```

After setting hostname and domain name, generate RSA key to start SSH service. 

```
crypto key generate rsa
```
