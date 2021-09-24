# HMS Discovery Internals
## Unified MEDS/REDS Vault Loader
> `*` does not include the RTS Vault loader

The HMS Discovery service will use its own space in vault, and not re-use/over-write values pushed in by REDS/MEDS. 

In addition the sealed secrets provided by MEDS and REDS will be reused. This will cause minimal changes when upgrade to the new discovery service. 

The Existing MEDS credential:
```bash
ncn# kubectl exec -it -n vault -c vault cray-vault-0  -- sh -c "export VAULT_ADDR=http://localhost:8200; vault kv get secret/meds-cred/global/ipmi"
====== Data ======
Key         Value
---         -----
Password    inital0
Username    root
```

Existing REDS credentials:
```bash
ncn-m001:~ # kubectl exec -it -n vault -c vault cray-vault-0  -- sh -c "export VAULT_ADDR=http://localhost:8200; vault kv get secret/reds-creds/defaults"
==== Data ====
Key     Value
---     -----
Cray    map[password:initial0 username:root]
command terminated with exit code 2
ncn-m001:~ # kubectl exec -it -n vault -c vault cray-vault-0  -- sh -c "export VAULT_ADDR=http://localhost:8200; vault kv get secret/reds-creds/switch_defaults"
========== Data ==========
Key                 Value
---                 -----
SNMPAuthPassword    testpass1
SNMPPrivPassword    testpass2
SNMPUsername        testuser
```

Potential new layout. All credentials used by hms-discovery are now under `secret/hms-discovery-creds` with saner names.
```
secret/hms-discovery-creds/mountain-bmc
    {
        "Username": "root",
        "Password": "password"
    }
secret/hms-discovery-creds/river-bmc
    {
        "Cray": {
            "Username": "root",
            "Password": "password"
        }
    }
secret/hms-discovery-creds/river-snmp
    {
        "SNMPUsername": "testuser",
        "SNMPAuthPassword": "authPassword",
        "SNMPPrivPassword": "privPassword"
    }
```

## Discovery of mountain hardware
### MAC Based discovery
__In CSM 1.1__ and before MEDS at startup would pre-populate HSM with every possible MAC address + Xname combination possible in a cabinet. It would then attempt to hit the redfish root for every potential redfish endpoint in the system. 

When it starts to ping for the first time is when it starts to 
  - Setup NTP/RSYSLog/SSHKeys on the BMC
  - Populate Vault with credentials for the BMC
  - Create/Update a Redfish Endpoint in HSM

The new HSM discovery service will instead discovery mountain in a similar fashion to how the River discovery process works 

### Powering up of mountain hardware to discovery Node/Router BMCs in a chassis
__TODO__

## MAC based Discovery flow
1. BMC requests a DHCP lease from KEA and then allocates an IP address for the BMC. Then KEA Populates (or updates the MAC address) `/v1/Inventory/EthernetInterfaces` with the BMC MAC address and the IP address it handed out
    > This is how KEA works in general for all devices that DHCP not just BMCs.

    > This is same process for both River and Mountain Hardware
    
    The following is a sample payload sent by KEA:
    ```json
    {
        "MACAddress": "08:f1:ea:80:3d:7f",
        "IPAddress": "10.254.1.27"
    }
    ```
    This will create a new EthernetInterface under HSM `/v2/Inventory/EthernetInterfaces/08f1ea803d7f`. The normalized MAC address is what is used to identify the MAC address in the HSM APIs.
    ```json
    GET /v2/Inventory/EthernetInterfaces/08f1ea803d7f
    {
        "ID": "08f1ea803d7f",
        "Description": "",
        "MACAddress": "08:f1:ea:80:3d:7f",
        "LastUpdate": "2021-06-28T18:18:15.960235Z",
        "ComponentID": "",
        "Type": "",
        "IPAddresses": [
            {
                "IPAddress": "10.254.1.27"
            }
        ]
    }
    ``` 
2. The MAC identification task in the discovery service looks for all MAC address that have an empty component ID. 
    > `GET /v2/Inventory/EthernetInterfaces?ComponentID`

3. Retrieve the default River and Mountain BMC credentials from Vault from:
    - `secret/hms-discovery-creds/river-bmc`
        ```json
        {
            "Cray": {
                "Username": "root",
                "Password": "password"
            }
        }
        ```
    - `secret/hms-discovery-creds/mountain-bmc`
        ```json
        {
            "Username": "root",
            "Password": "password"
        }
        ```
    - `secret/hms-discovery-creds/river-snmp`
        ```json
        {
            "SNMPUsername": "testuser",
            "SNMPAuthPassword": "authPassword",
            "SNMPPrivPassword": "privPassword"
        }
        ```

4. Query SLS for known management switches
    
    For each found management switch perform the following:
    1. Query Vault the SNMP credentials for the switch at `secret/hms-creds/<xname>`
        ```json
        {
            "Password": "",
            "SNMPAuthPass": "testpass1",
            "SNMPPrivPass": "testpass2",
            "URL": "",
            "Username": "testuser",
            "Xname": "x3000c0w14"
        }
        ```

    1. If no credentials exist at that location then check SLS to see if the Passwords exist there in plain text
        > TODO: Should this "feature" be removed for security purposes, as all of our passwords should be stored and coming from vault.
    
    1. If no credentials exist in Vault, and are not present in SLS the push in the default SNMP credentials into Vault at `secret/hms-creds/<xname>`.
        __TODO__:
        ```json
        {
            "Xname": "x3000c0w14",
            "Username": "testuser",
            "SNMPAuthPass": "testpass1",
            "SNMPPrivPass": "testpass2"
        }
        ```


3. Identification of the unknown MAC address 
    1. __Mountain BMC Handling__: For each unknown MAC address perform the following check:
        > The MAC address format for Mountain BMCs
        > ```
        > 00:00:00:00:00:00
        > || || || || || ||
        > || || || || || |\- Controller defined 'base' : [0x00 - 0x0F]
        > || || || || || \-- Sub component 'index'     : [0x00 - 0x0F]
        > || || || || \----- Slot + offset             : [0x00 - 0xFF]
        > || || || \-------- Chassis                   : [0x00 - 0xFF]
        > || \-------------- Rack                      : [0x0000 - 0xFFFF]
        > \----------------- MAC pool prefix
        > ```
        1. Does the MAC have the MAC pool prefix of `02`? If it does its a mountain BMC, other wise it is River BMC
        2. From the MAC address determine what this device is.
            ```go
            // Break apart the MAC address into its components
            macPoolPrefix := macAddress[0]
            rack := int32(macAddress[1]) << 8 + int32(macAddress[2])
            chassis := int32(macAddress[3])
            slotAndOffset := int32(macAddress[4])
            subComponentIndex := int32(macAddress[5] & 0xf0) >> 4
            controllerBase := int32(macAddress[5] & 0x0f)
            ```

            The `slotAndOffset` value can determine why type of device this is:
            - `0 <= slotAndOffset < 48` are ChassisBMCs
            - `48 <= slotAndOffset < 96` are NodeBMCs
            - `96 <= slotAndOffset < 144` are RouterBMCs
            - `144 <= slotAndOffset < 256` are Blade controllers

            Currently we do not support discovering Blade controllers, and any MAC that is a blade controller will be ignored. Also with current hardware blade controllers do not show up on the HMN network.

            Then the xname can be built up by using the broken up MAC address:
            ```go
            // Build up the XName
            var xname string
            switch(deviceType) {
            case base.ChassisBMC:
                // xXcCbB
                xname = fmt.Sprintf("x%dc%db%d", rack, chassis, subComponentIndex) 
            case base.NodeBMC:
                // xXcCsSbB
                xname = fmt.Sprintf("x%dc%ds%db%d", rack, chassis, slot, subComponentIndex) 
            case base.RouterBMC:
                // xXcCrRbB
                xname = fmt.Sprintf("x%dc%dr%db%d", rack, chassis, slot, subComponentIndex) 
            default:
                return "", fmt.Errorf("unknown device type: %s", deviceType)
            }
            ```

        3. If a valid xname is able to built for the MAC address continue the next section, otherwise try to identify using the River BMC handling procedure. 

    2. __River BMC Handling__
        1. For each leaf switch perform the following SNMP queries to build up a MAC address to switch port map. This will result in a data structure that looks like: `map[switch xname]map[mac address][switch port]`
            1. Bulk Get on `1.3.6.1.2.1.31.1.1.1.1` for Port map.
            1. Bulk Get on `1.3.6.1.2.1.17.1.4.1.2` for Port number map.
            1. Bulk Get on `1.3.6.1.2.1.17.7.1.2.2.1.2` for non-VLAN MAC port map.
            1. Bulk Get on `1.3.6.1.2.1.17.4.3.1.2` for VLAN MAC port map.


            For Aruba switches the switch port map will look simular to:
            ```json
            {
                "x3000c0w14": {
                    "6805cabbc182": "1/1/14",
                    "9440c936bb36": "1/1/28",
                    "9440c937671c": "1/1/27",
                    "9440c9376756": "1/1/30",
                    "9440c9376776": "1/1/29",
                    "9440c9377700": "1/1/31"
                }, 
                "x3000c0w15": {
                    "9440c937d3e6": "1/1/26",
                    "9440c937e344": "1/1/25",
                    "9440c937f382": "1/1/37",
                    "9440c937f3f2": "1/1/48",
                    "9440c93808c7": "1/1/46"
                }
            }
            ```

            For Dell switches the switch port map will look similar to:
            ```json
            {
                "x3000c0w14": {
                    "6805cabbc182": "ethernet1/1/14",
                    "9440c936bb36": "ethernet1/1/28",
                    "9440c937671c": "ethernet1/1/27",
                    "9440c9376756": "ethernet1/1/30",
                    "9440c9376776": "ethernet1/1/29",
                    "9440c9377700": "ethernet1/1/31"
                }, 
                "x3000c0w15": {
                    "9440c937d3e6": "ethernet1/1/26",
                    "9440c937e344": "ethernet1/1/25",
                    "9440c937f382": "ethernet1/1/37",
                    "9440c937f3f2": "ethernet1/1/48",
                    "9440c93808c7": "ethernet1/1/46"
                }
            }
            ```

        2. For each unknown MAC address check it against the switch port map. Then search through the each switches MAC Address to switch port map.
            1. If the MAC address was not found for a certain switch, then look at the next switch MAC port mapping.

            1. Query SLS with the switch xname and port combination to determine the device connected to this port.
                > Lookup switch port `1/1/14` on switch `x3000c0w14`
                ```json
                GET /v1/search/hardware?type=comptype_mgmt_switch_connector&class=River&extra_properties.VendorName=1/1/14&parent=x3000c0w14

                [
                    {
                        "Parent": "x3000c0w14",
                        "Xname": "x3000c0w14j14",
                        "Type": "comptype_mgmt_switch_connector",
                        "Class": "River",
                        "TypeString": "MgmtSwitchConnector",
                        "LastUpdated": 1624396512,
                        "LastUpdatedTime": "2021-06-22 21:15:12.728105 +0000 +0000",
                        "ExtraProperties": {
                            "NodeNics": [
                                "x3000c0s9b0"
                            ],
                            "VendorName": "1/1/14"
                        }
                    }
                ]
                ```

                If `null` returned by SLS then this Switch and Switch port combination is not known to SLS. then the look at the next switch port mapping.


                The BMC/XName that is connected to this port can be found in the `NodeNics` array.
                > Right now there should only be 1 element in NodeNics, as it is expected that only 1 BMC is connected per switch port.
                > 
                > In the future when Apollo 2000 need to get supported this will need to change. Apollo 2000s are a 2U dense compute node chassis with 4 compute nodes. The 4 BMCs in Apollo 2000 chassis will not have a dedicated connection to the HMN, and instead the 4 BMCs will make a single connection to the HMN via a Rack Consolidation Module (RCM). The RCM is like a dumb switch, so all 4 MAC address will show up under the a single switch port. So an additional step will be required to discovery these BMCs to determine which of the 4 BMCs the BMC mac address is associated with.
                > 
                > This require changes in the following changes:
                > 1. When CSI encounters a HMN row with `RCM` in the name it will need to build up the `MgmtSwitchConnector` connector differently, then a standard node.
                >   1. CSI will not generate MgmtSwitchConnector for anything that has a parent with `RCM` in the xname
                >   2. When building a `MgmtSwitchConnector` for the RCM it should create a `MgmtSwitchConnector` for all of the hardware that the RCM parents.
                >   > All of the BMC xnames in NodeNics will have the name `xXcCsS`, and the `bB` will be the only differing thing.
                >   ```json
                >   {
                >       "Parent": "x3000c0w14",
                >       "Xname": "x3000c0w14j14",
                >       "Type": "comptype_mgmt_switch_connector",
                >       "Class": "River",
                >       "TypeString": "MgmtSwitchConnector",
                >       "ExtraProperties": {
                >           "NodeNics": [
                >               "x3000c0s9b1"
                >               "x3000c0s9b2"
                >               "x3000c0s9b3"
                >               "x3000c0s9b4"
                >           ],
                >           "VendorName": "1/1/14"
                >       }
                >   }
                >   ```
                > 2. When the discovery job finds a `MgmtSwitchConnector` with multiple `NodeNics`, it needs to interrogate the BMC to determine which of the 4 it actually is.
                >   1. Query Redfish on the BMC to figure out the bay number:
                >   ```bash
                >   ncn# curl -s -k -u user:password https://a2000n1-bmc/redfish/v1/Chassis/1/ | jq .Oem.Hpe.BayNumber
                >   1
                >   ncn# curl -s -k -u user:password https://a2000n2-bmc/redfish/v1/Chassis/1/ | jq .Oem.Hpe.BayNumber
                >   2
                >   ncn# curl -s -k -u user:password https://a2000n3-bmc/redfish/v1/Chassis/1/ | jq .Oem.Hpe.BayNumber
                >   3
                >   ncn# curl -s -k -u user:password https://a2000n4-bmc/redfish/v1/Chassis/1/ | jq .Oem.Hpe.BayNumber
                >   4
                >   ```                
                >   2. The Bay number will be the `B` in `xXcCsSbB` for one of the Node NICs xname in the MgmtSwitchConnector. This is the Xname for this BMC.
            
            3. If the MAC address was not found in any switch port map, then the MAC address is still unknown.
        
    3. __Unknown device handling__: Any MAC address that its identified cannot be determined will be ignored, and no further action will be taken with it.

3. Ping the device to see if it speaks Redfish:

    Using the IP address perform a HTTP get to the Redfish Service root `/v1/redfish/`
    - If the BMC responds 200 then proceed.
    - If this is a CabinetPDUController, then provided.
    - Otherwise to not proceed further, and do not update Hardware State Manager.

    > 1. If the BMC is not in a healthy state the DNS name for the BMC will not
    > Some concerns:
    >    be available. Will have to use a manual process to determine the IP of the
    >    the BMC.
    >      - Need to be careful of Nodes on the NMN that the discovery service could 
    >        mis-identify as a BMC. Example of this would be Gazelle.
    >      - Could provide a script to identify potential IP address of the unhealthy 
    >     BMC to ease the trouble shooting process when the hms-discovery-verify 
    >     script shows that that BMC is not healthy.
    > 2. Update the MAC address in HSM. Then Query EthernetInterfaces for all NodeBMC/
    >    Chassis/CabinetPDUControllers and check to see if all of them are present in 
    >    HSM as a RedfishEndpoint. If not them create the redfish endpoint.
    >    - Concerned about the load put onto HSM

4. If this device is a CabinetPDUController `xXmM`:
    > There are 2 types of PDUs that can be found on a Shasta system ServerTech and HPE PDUs. 
    > 1. HPE PDUs support Redfish, and should be inventoried via HSM like any other BMC
    > 2. ServerTech PDUs do not speak Redfish, and require the use of Redfish Translation Service (RTS) for HMS services to interact with the PDU via Redfish.

    1. Check to see if the PDU speaks Redfish using the result of the step above
        - If it speaks Redfish continue to the next section.
            > HPE PDUs need to have Redfish enabled on them.
            ```
            ncn-m001:~ # curl -k -i  https://x3000m0/redfish/v1/
            HTTP/1.1 500 Internal Server Error
            Server: HPE_PDU_GEN2/1.4.0
            Content-type: application/json
            Connection: keep-alive
            Content-Length: 54

            {"ErrorDescription": "Redfish service is not running"}
            ```

        - If it does not speak Redfish, but has `ServerTech` in the headers from the response body above then continue with this section.
            ```bash
            ncn-m001:~ # curl -k -i  https://x3000m0/redfish/v1/
            HTTP/1.1 303 See Other
            Location: https://x3000m0/
            Content-Length: 0
            Server: ServerTech-AWS/v8.0p
            ```

    2. Get the default PDU credentials from Vault. This are populated in Vault by RTS.
        ```bash
        ncn-m001:~ # kubectl exec -it -n vault -c vault cray-vault-0  -- sh -c "export VAULT_ADDR=http://localhost:8200; vault kv get secret/pdu-creds/global/pdu"
        ====== Data ======
        Key         Value
        ---         -----
        Password    password
        Username    username
        ```
        If no default credentials exist, then stop processing this MAC address. 
          > This allows for the scenario where RTS has not fully been initialized or RTS is not installed but the hms-discovery-service is present.

    2. Update the MAC address in Hardware State Manager to include the PDU Xname.

        Perform a PATCH to `/v2/Inventory/EthernetInterfaces/<MAC>`:
        ```json
        {
            "ComponentID": "<BMC XNAME>"
        }
        ```

    3. Store a new PDU Credential into Vault at `secret/pdu-creds/<xname>`, where `xname` is in the form of `x3000m0`
        ```json
        secret/pdu-creds/<xname>
        {
            "Xname": "<xname>",
            "Username": "<defaultUsername>",
            "Password": "<defaultPassword>",
            "URL": "https://<xname>/jaws"
        }
        ```

    4. No more work is required by the HSM discovery service to get the PDU discovered, as the rest will be handled by RTS.


1. Configure RSYSLOG/NTP/SSH Keys on BMC that support it:
    
    The following BMCs support RSYSLOG/NTP/SSH Keys:
    - Mountain and Hill
        - ChassisBMC
        - NodeBMC
        - RouterBMC
    - River
        - RouterBMC

    The `Class` of the device can be determined by looking at the method that was used to identify the MAC address.

    1. Retrieve SSH BMC Credentials
    __TODO__ This needs to explored further

    1. Configure NTP/SSH Keys/RSYSLOG the BMC with a PATCH to `https://<BMC_IP>/redfish/v1/Managers/BMC/NetworkProtocol`
        > The hostname for the NTP server will be converted to an IP address in the hms-bmc-network-protocol library.

        > If no SSH keys are available, then they will not be included in the payload. These are not currently being set with MEDS.
        ```json
        {
            "Oem": {
                "Syslog": {
                    "ProtocolEnabled": true,
                    "SyslogServers": [
                        "rsyslog-aggregator.hmnlb"
                    ],
                    "Transport": "udp",
                    "Port": 514
                },
                "SSHAdmin": {
                    "AuthorizedKeys": "<SSH KEY>"
                },
                "SSHConsole": {
                    "AuthorizedKeys": "<SSH KEY>"
                }
            },
            "NTP": {
                "NTPServers": [
                    "1.1.1.1"
                ],
                "ProtocolEnabled": true,
                "Port": 123
            }
        }
        ```


4. Store the BMC credentials into Vault under `secret/hms-creds/<xname>`.
    > This makes the BMC credentials available to other services like HSM, CAPMC, etc...

    ```json
    {
        "Xname": "<xname>",
        "Username": "<defaultRiverUsername>",
        "Password": "<defaultRiverPassword>"
    }
    ```


3. Update the MAC address in Hardware State Manager with the correct ComponentID
    Perform a patch to `/v2/Inventory/EthernetInterfaces/<MAC>`:
    ```json
    {
        "ComponentID": "<BMC XNAME>"
    }
    ```

5. Tell HSM to perform an inventory on the BMC:

    Attempt to `POST` the following payload to HSM at `/v2/Inventory/RedfishEndpoints`
    ```json
    {
        "ID": "<xname>",
        "FQDN": "<fqdn>",
        "MACAddr": "<mac>",
        "RediscoverOnUpdate": true,
        "Enabled": true
    }
    ```

    If a `409` status code is returned, then attempt to PATCH the existing RedfishEndpoint at `/v2/Inventory/RedfishEndpoints/<xname>` with the same payload.

## Task: Populate HSM with NCNs
__TODO__

## Random Thoughts
- The Discovery job should check to see if a EthernetInterface has a RedfishEndpoint. If it is reachable, then readd the RedfishEndpoint. This will make the behvaior simular to MEDS.
- Bring back the concept of the MEDS ping. It can be used as a mechanism to automatically bring back in RedfishEndpoints into HSM.
    - Alternative, for every BMC MAC that has a IP + component ID check with HSM to see if that redfish Endpoint exists
        - If it does not, re-add it.