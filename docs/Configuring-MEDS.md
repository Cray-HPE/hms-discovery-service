1.  [CASMHMS](index.html)
2.  [CASMHMS Home](CASMHMS-Home_119901124.html)
3.  [Design Documents](Design-Documents_127906417.html)
4.  [Discovery](Discovery_227479292.html)

# <span id="title-text"> CASMHMS : Configuring MEDS </span>

Created by <span class="author"> Steven Presser</span>, last modified on
May 02, 2019

Configuring MEDS is a simple process.  Meds takes 2 pieces of
information as inputs:

-   The IP prefix the cabinets are configured to use
-   The numbers of the racks in the system

# Summary

To configure MEDS, override the following variables in ansible:

-   cray\_meds\_rack\_numbers – A list of rack numbers present in the
    system.  Default: \[ \] (an empty list)

-   cray\_meds\_ip\_prefix – The IP prefix configured for use in the
    system.  Default fd66:0:0:0

# Configuring on an Installed System

To configure MEDS on an installed system, edit
/root/k8s/cray\_meds.yaml.  The default file looks something like this:

    ---
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: cray-meds
      labels:
        app: cray-meds
    spec:
      template:
        metadata:
          labels:
            app: cray-meds
            api-gateway: upstream
        spec:
          volumes:
          - name: ca-vol
            configMap:
              name: cray-configmap-ca-public-key
          - name: admin-client-auth
            secret:
              secretName: "{{ kong_admin_client_auth_secret_name }}"
          containers:
          - name: cray-meds
            image: sms.local:5000/cray/cray-meds:latest
            env:
            - name: HSM_URL
              value: https://api-gateway.default.svc.cluster.local/apis/smd/hsm/v1
            - name: MEDS_OPTS
              value: "-ipprefix=fd66:0:0:0 -rack 3 -rackip 10.100.106.2/22"
          hostname: meds
          nodeSelector:
            bmc-net: "True"
          hostNetwork: true

To change the options in use , edit the MEDS\_OPTS value line.  To
change the ip prefix in use, change the value of `-ipprefix=`.  To add
racks, add `-rack RACKNUM` options, one per rack.  If you want to do
IPv4 discovery of racks, add -rackip options, one per rack.  The number
of -rackip options must be 0 or equal to the number of -rack options. 
The -rackip options will be applied in the same order as the -rack
options.  So the first -rackip attaches to the first -rack, the second
to the second, etc.

# Configuring pre-install (non-robot-managed system)<span id="ConfiguringMEDS-nonrobbot-config" class="confluence-anchor-link"></span>

Follow standard pre-installation ansible variable customization.  This
will have you modifying `customer_var.yml`.  You'll want to add values
for the "`cray_meds_racks`" and "`cray_meds_ip_prefix`" variables. 
These may end up looking like this:

    cray_meds_ip_prefix: fd66:0:0:0
    cray_meds_racks:
      - { "number": 1, "ip4net": "10.100.106.2/22" }
      - { "number": 2, "ip4net": "10.100.108.2/22" }
      - { "number": 3, "ip4net": "10.100.110.2/23" }
      - { "number": 5, "ip4net": "10.254.100.2/22" }

For more information, see standard install documentation.

Note that when not using IPv4 discovery, the "ip4net" parameter should
be omitted.

# Configuring pre-install (robot-managed system)

You will have to modify the `customer_var.yml` file stored in the
<a href="https://stash.us.cray.com/projects/SMTEST/repos/robot/browse/configs/" class="external-link">SMTEST robot repo</a>. 
It should be modified just like
in [nonrobot-config](#ConfiguringMEDS-nonrobot-config).

## Comments:

<table data-border="0" width="100%">
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><span id="comment-138021319"></span>
<p><a href="https://connect.us.cray.com/confluence/display/~spresser" class="confluence-userlink user-mention">Steven Presser</a> It would be good to document configuring MEDS IPV4</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by jrouault at May 02, 2019 10:17
</div></td>
</tr>
<tr class="even">
<td style="border-top: 1px dashed #666666"><span id="comment-138021463"></span>
<p>Agreed.  It's on my to-do list, but I want to get the functionality in the repo before I document it as if it can be done.</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by spresser at May 02, 2019 12:21
</div></td>
</tr>
</tbody>
</table>

Document generated by Confluence on Jan 14, 2022 07:17

[Atlassian](http://www.atlassian.com/)
