1.  [CASMHMS](index.html)
2.  [CASMHMS Home](CASMHMS-Home_119901124.html)
3.  [Design Documents](Design-Documents_127906417.html)
4.  [Discovery](Discovery_227479292.html)

# <span id="title-text"> CASMHMS : REDS/MEDS/SLS </span>

Created by <span class="author"> Rod Frost</span>, last modified by
<span class="editor"> Ryan Haasken</span> on Aug 15, 2019

# Tasks

The REDS/MEDS/SLSstory involves several specific tasks:

-   Endpoint discovery: This is the process of finding hardware on the
    network
-   Inventory discovery: The process of determining what capabilities
    the hardware has (disk, memory, cores, etc)
-   Geolocation: The process of assigning an xname to hardware based on
    its physical location.

# SLS

SLS(System Layout Service) serves as a “single source of truth” for the
system design. It details the physical locations of network hardware,
compute nodes and cabinets. Further, it stores information about the
network, such as which port on which switch should be connected to each
compute node. This information is critical to Geolocation in River
hardware. (Mountain hardware, in contrast, is able to perform
geolocation without requiring such a detailed mapping because Cray has
built this functionality into the hardware). SLS does not store the
details of the actual hardware (e.g.: hardware identifiers). Instead it
stores a generalized abstraction of the system other services may use.
SLS thus does not need to change as hardware within the system is
replaced.

SLS presents a simple to use HTTP API for querying the stored
information.

Cray does not anticipate customers needing to interact with SLS
frequently or in a significant manner, unless making changes to system
hardware that were not planned for at system creation. However,
interaction with SLS is required if the system setup changes – for
example, if system cabling is altered or if the system is expanded or
reduced.  SLS does not interact with the rest of the system. 
Interaction with SLS should occur only during system installation,
expansion and contraction.

SLS uses site.yaml and other SCT output as it's sources of information. 
It is responsible for:

-   Providing an HTTP API to access this site information
-   At startup or when objects are created, for notifying HSM/smd of
    endpoints that are currently missing and marking these as "empty"
-   Storing a list of all hardware (Mountain CEC, Rack, cC, nC, sC,
    river nodes, PDUs)
-   Storing a list of all network links
-   Storing a list of all power links (eg: PDU port to node in River)
-   Adding any non-discoverable hardware (eg: river PDUs) to HSM/smd if
    not already present

  

# REDS

Design document: [River Endpoint Discover Service (HMS)](110180596.html)

The River Endpoint Discovery Service (REDS) manages Endpoint Discovery,
Initialization, and Geolocation of River hardware. Because it interacts
with commodity off-the-shelf (COTS) hardware, it is relatively more
complex than the other services. REDS only interacts with the rest of
the system when an “unknown” compute node attempts to boot on the
network. A compute node is unknown if it is not already listed in
hardware state manager. Customers should have to interact with REDS only
to correct error conditions.

REDS consists of two major components: a daemon hosted in the management
plane and a discovery image that is booted on compute nodes. These two
cooperate to gather core information about compute nodes and network
addresses. This core information is then used to identify the node and
send information about its BMC to HSM (which then performs detailed
hardware discovery).

First, the discovery image. The image is a standard PXE boot image. When
a node attempts to PXE boot and is not known to HSM, it is told to boot
the discovery image. This image contacts the REDS daemon to get
configuration information (such as authentication credentials for the
BMC) and configures the BMC. It then gathers information about the BMC
(such as the IP address assigned to the BMC and the MAC address of the
BMC interfaces) and sends this to the daemon.

The daemon acts as a coordinator, taking in information from multiple
sources and using it to perform identification of the nodes. Broadly
speaking, the daemon associates an IP address to an xname (and thus a
location) by associating the IP address to a MAC address in the
discovery image. It then determines which switch and switch port that
MAC address is connected to. Finally, it queries the switch and switch
port in SLS to determine what xname is connect to that port. Note that
interaction between REDS and the switches happens asynchronously from
interaction between REDS and the compute node.

The daemon has two major components: an HTTP module to handle
interactions with compute nodes and an SNMP module to handle interaction
with network switches. The HTTP module exists to handle information flow
to and from the discovery image booted on the compute nodes. The SNMP
module gathers information on which MAC address is plugged into which
port on each switch. This information is combined with the MAC address
discovered in the discovery image to allow a bridge between the network
layout and the physical layout of the system. The SNMP module also
determines when a node is no longer available on the network by when it
disappears from the switch it is connected to. The core of the daemon
combines the information from the HTTP and SNMP modules. Once the daemon
has an IP address, MAC address, username and password for a BMC, it
notifies HSM that a new endpoint exists. In the case of a disappearing
node, it determines which xname matches the MAC address that has
disappeared, then notifies HSM that the node is no longer present.

<table class="gliffy-macro-table" width="100%">
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><table class="gliffy-macro-inner-table">
<tbody>
<tr class="odd">
<td><img src="attachments/135301926/135301936.png" class="gliffy-macro-image" /></td>
</tr>
</tbody>
</table></td>
</tr>
</tbody>
</table>

# CCD Reader

New systems are starting to be built with CCDs in the approved format
for Shasta.  A CCD is an excel spreadsheet that contains all the
information necessary for manufacturing to build and configure the
system.  It is a wealth of information on many topics.

REDS and HSM each have an input file of information about the system. 
For REDS this is the REDS mapping file, which contains a wiring map of
the system.  For HSM it is the node\_nid\_map file, which assigns roles
and nids to the nodes within the system.  Both of these files have
previously been built by hand.  However, with information appearing in
the CCD, these can be largely autogenerated.  A tool to do so is
available
in <https://stash.us.cray.com/projects/HMS/repos/hms-ccd-reader/browse>.

When using the tool, it will read the CCD and extract the information it
is able.  It will then prompt for the remaining information about the
system needed to fill out both files.  The files then become input to
the install process.

# MEDS Current

The Mountain Endpoint Discovery Service (MEDS) manages Endpoint
Discovery, Initialization, and Geolocation of Mountain Hardware.  The
Mountain Endpoint Discovery Service (MEDS) is an extremely simple
service. Shasta Mountain hardware assigns itself a hostname, MAC and IP
address based on its location, which it is able to determine by reading
from an embedded cabinet controller.  In the current implementation of
MEDS the MEDS service is told what racks are in the system and the IP
address ranges assigned to each.  From the rack numbers an IP ranges, it
algorithmically calculates the IP addresses that should be assigned to
each piece of hardware in the system.  It then periodically makes
Redfish requests of these endpoints to determine if they are present or
not.  When nodes first become present, they are added to HSM.

# MEDS FUTURE

Future improvements to MEDS include:

-   Noting hardware as "disabled" in HSM when it leaves the network
-   Noting hardware as "enabled" in HSM when it re-appears and
    restarting inventory discovery
-   Reading Redfish endpoints fo chassis controllers to determine if a
    slot is empty or merely powered off.

## Comments:

<table data-border="0" width="100%">
<colgroup>
<col style="width: 100%" />
</colgroup>
<tbody>
<tr class="odd">
<td><span id="comment-149904316"></span>
<p>Is SLS a Level 2 Service (Infrastructure as a Service)?</p>
<div class="smallfont" data-align="left" style="color: #666666; width: 98%; margin-bottom: 10px;">
<img src="images/icons/contenttypes/comment_16.png" width="16" height="16" /> Posted by ekoen at Sep 27, 2019 12:32
</div></td>
</tr>
</tbody>
</table>

Document generated by Confluence on Jan 14, 2022 07:17

[Atlassian](http://www.atlassian.com/)
