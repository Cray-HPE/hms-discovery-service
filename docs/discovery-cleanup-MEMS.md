1.  [CASMHMS](index.html)
2.  [CASMHMS Home](CASMHMS-Home_119901124.html)
3.  [Design Documents](Design-Documents_127906417.html)
4.  [Discovery](Discovery_227479292.html)

# <span id="title-text"> CASMHMS : Discovery Cleanup (Remove REDS/MEDS, create MEMS) </span>

Created by <span class="author"> Steven Presser</span> on Feb 19, 2021

As the Shasta system has evolved, the discovery services have not
necessarily kept up. They've kept working, sometimes via quickly-written
functionality. It's time to clean these up and draw clear distinctions
in their roles.

# Current Services

This section will walk through the current services and their current
responsibilities.

## REDS (River Endpoint Discovery Service)

The most venerable of these services. REDS originally worked via booting
an image on nodes to configure them. Over time, this became unnecessary.

REDS is now responsible solely for configuring NTP and syslog on rosetta
switches when they are added to the system.

## MEDS (Mountain Endpoint Discovery Service)

MEDS has changed remarkably little since inception. It queries SLS for
Mountain cabinets, generates a list of endpoints, then shoves those into
smd's ethernetInterfaces table for use by hms-discovery.

When those endpoints become available, it is then responsible for adding
them to smd. It checks this via a "redfish ping", where it attempts to
load the root of the redfish tree to verify the endpoint is reachable.

## hms-discovery

`hms-discovery` is responsible for discovering River nodes of all types
(including determining the xname of the hardware), and for powering on
Mountain hardware so that MEDS can discover it.

# Problems

There are several problems with this as it stands:

-   More than one service performs NTP/syslog configuration of
    RedfishEndpoints. Since all Redfish endpoints that can be configured
    with this are Cray hardware, it simply doesn't make sense to have
    the same code called two different ways and with two sources of
    configuration data.
-   Many tasks related to hardware are done just once at service
    startup. This includes "stuffing" smd's `ethernetInterfaces` table
    with the MACs of mountain hardware and adding NCNs and river
    rosettas to HSM. This means that these tasks are not independently
    rerunable, nor do they automatically update when new hardware is
    added to the system or SLS is updated.
-   MEDS has too much code that's tied to proscribed IP addresses. (A
    now-removed feature)
-   MEDS comes up with and stuffs MAC addresses for all possible
    hardware, even on cabinets that are less-than-fully populated. This
    can result in meaningless entries and make it more difficult to find
    useful information in the `ethernetInterfaces` table
-   Nothing routinely verifies that the NTP/syslog and power
    subscription settings are correct on redfish endpoints

# The Proposal

We recently removed all code that added and removed redfishEndpoints
when compute nodes enter or leave the system. This means that from a
discovery standpoint we now have two major events in a node's lifetime:

1.  When the node is first added to the system
2.  When the node is re-added to the system (eg: after maintenance)

Additionally, we have occasional circumstances where we need to
reconfigure nodes. For example, Mountain nodes, if they are unable to
notify subscribers of power state changes, will stop attempting to do
so. Or a River node may require `ipmi mc reset cold`, which breaks
redfish subscriptions, resulting in hsm not knowing when the node powers
up or down.

I propose we split the discovery role into two parts:

1.  First discovery. When the node is added to the system for the first
    time and must have certain tasks done, including associating the mac
    address to a node name. These are one-time-only tasks which largely
    have to do with smd state and should not be repeated.
2.  Discovery configuration. These tasks should be done whenever a node
    may have been altered or reset. This includes syslong/NTP
    configuration, subscription information and updating HSM's node
    inventory.

Fortunately, two of our pieces of software mostly fulfill these two
roles already. `hms-discovery` excels at first discovery for
River-conencted nodes (including NCNs). MEDS excels at monitoring
redfishEndpoints and triggering rediscovery, which are tasks much like
discovery configuration.

I propose the following work:

1.  The responsibility for "stuffing" ethernetInterfaces with Mountain
    data be removed from MEDS. This could be placed in a k8s job to run
    on demand. Alternatively, `hms-discovery` could be given the
    responsibility of noting new `ethernetInterface` entries with
    Mountain MAC addresses and determining the xname for them. The first
    option is less work, while the second results in fewer unused
    entries in the `ethernetInterfaces` table.

2.  MEDS be renamed as the "Management Endpoint Monitoring Service"
    (MEMS) (that is, move the codebase and give it a new name). This new
    service has two roles:

    First, to monitor redfishEndpoints listed in HSM. Whenever one
    changes state from "not accessible on the network" to "accessible on
    the network" of from "disabled in HSM to "enabled in HSM", MEDS-next
    shall reconfigure any syslog/NTP configuration and instruct HSM to
    rediscover it (so as to reestablish HSM subscriptions and pick up
    any hardware updates). These updates to HSM may be batched so as not
    to overload HSM.

    Second, MEDS-next will be responsible for periodic reconfiguration
    of nodes. Periodic rediscovery shall consist of the same tasks as in
    monitoring redfishEndpoints. This shall be kept at a low frequency
    so as not to cause undue load on HSM or the nodes. This periodic
    reconfiguration will ensure the correct settings are present on the
    node if they have been lost for whatever reason. Specifically,
    periodic rediscovery will resolve nodes that have stopped sending
    subscription notifications.

    I believe it is best to put both these tasks in the same software
    because both involve configuring the redfishEndpoints. Doing these
    two tasks in one piece of software ensures that there is only one
    central place where these tasks are performed and that the result is
    the same, no matter the cause of needing to reconfigure the node.

3.  Responsibility for configuring NTP/syslog on River Rosettas is moved
    to MEMs

4.  REDS, now devoid of purpose, be removed entirely.

Changing to these roles will give the following advantages:

-   Each service/job has clearly defined
    responsibilities. `hms-discovery` is responsible for getting
    redfishEndpoints added to HSM. MEDS-next is responsible for ensuring
    nodes are configured with the correct syslog/NTP and subscription
    settings.
-   Each job/service is kept as simple as possible
-   The number of disparate services/jobs is reduced
-   Jobs perform defined tasks and administrators can run them
    on-demand. Restarting services will no longer cause changes in state
    on the system.
-   To the greatest extent possible, dependency on particular hardware
    types is minimized

  

Document generated by Confluence on Jan 14, 2022 07:17

[Atlassian](http://www.atlassian.com/)
