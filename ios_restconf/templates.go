package ios_restconf

const tenantTemplate = `
{
  "ned:native": {
    "ip": {
      "vrf": [
        {
          "name": "{{ .VRF }}"
        }
      ],
      "nat": {
        "inside": {
          "source": {
            "list": [
              {
                "id": "NAT",
                "interface": {
                  "{{ .OutsideInterfaceType }}": "{{ .OutsideInterfaceID }}"
                },
                "vrf": "{{ .VRF }}",
                "overload": [
                  null
                ]
              }
            ]
          }
        }
      }
    },
    "interface": {
      "GigabitEthernet": [
        {
          "name": "{{ .InsideInterfaceID }}.{{ .VlanID }}",
          "encapsulation": {
            "dot1Q": {
              "vlan-id": {{ .VlanID }}
            }
          },
          "ip-vrf": {
            "ip": {
              "vrf": {
                "forwarding": "{{ .VRF }}"
              }
            }
          },
          "ip": {
            "address": {
              "primary": {
                "address": "{{ .Gateway }}",
                "mask": "255.255.255.0"
              }
            },
            "policy": {
              "route-map": "NAT"
             },
            "nat": {
              "inside": [
                null
              ]
            }
          }
        }
      ]
    }
  }
}
`

const floatingIPTemplate = `
{
  "ned:nat": {
    "inside": {
      "source": {
        "static": {
          "nat-static-transport-list": [
            {
              "local-ip": "{{ .LocalIP }}",
              "global-ip": "{{ .GlobalIP }}",
              "vrf": "{{ .VRF }}"
            }
          ]
        }
      }
    }
  }
}
`
