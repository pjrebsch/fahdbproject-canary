
# Folding@Home Database Project Canary

The canary application runs on volunteer Folding@Home contributors' folding
machines to collect and report folding performance data for the Folding@Home
Database Project. It is meant to run in the background and require as little
maintenance for the volunteer folder as possible.

## Overview

The application performs read-only operations with the Folding@Home version
7 client and the host machine. Specifically, the types of info that are
collected and reported or used in some way by the canary are:

- PCI Graphics Device Information for each folding GPU (retrieved by one of various possible methods)
  - Vendor ID which corresponds to the GPU vendor (e.g. NVIDIA, AMD)
  - Device ID which corresponds to the GPU model (e.g. GF104 [GeForce GTX 460], Bonaire XT [Radeon HD 7790], etc.)
  - Subsystem Vendor ID if it can be found which corresponds to the GPU manufacturer (e.g. MSI, Gigabyte, Asus, etc.)
  - Subsystem Device ID if it can be found which corresponds to the GPU manufacturer's model
- GPU Device Driver Information (retrieved by one of various possible methods)
  - Driver Version
- GPU Slot Information (retrieved from the FAHClient using the "slot-info" telnet command)
  - GPU Index ("options":"gpu-index")
- Work Unit Information (retrieved from the FAHClient using the "queue-info" telnet command)
  - Project ("project")
  - Run ("run")
  - Clone ("clone")
  - Generation ("gen")
  - Folding@Home Core ("core")
  - Unit ID ("unit")
  - Total Frames for the work unit ("totalframes")
  - Number of frames observed (calculated value based on the "framesdone" attribute)
  - Time per frame of each frame observed (calculated value based on several attributes)
  - Work Unit Slot ("slot") to associate the work unit to a folding slot to ultimately associate it to a folding device
- FAHClient Information
  - Current version (e.g. "7.4.4")
  - Build platform (e.g. "linux2 3.2.0-1-amd64")
  - Build bits (e.g. 32, 64)
  - Current OS (e.g. "Linux 3.16.0-4-amd64 x86_64")
  - Current OS Architecture (e.g. AMD64)
- Folding@Home Database Project Canary Information
  - Current Canary version
  - Host platform (e.g. Windows, Linux)

## Wishlist

The primary goal of the Folding@Home Database Project is to collect and
analyze GPU folding performance, so this wishlist is ordered according to
the priorities of the project.

- At some point it would be nice to be able to get CPU utilization while
observing GPU work unit frames in order to determine if CPU utilization
is affecting the GPU's ability to perform as some GPU FAHCores require
a full CPU core to perform their work.
- It would likely also be helpful to collect and report on CPU performances
once GPU performance analysis techniques have become stable.

## Mechanics

### Obtaining PCI Graphics Device Information

In order to query information for the proper PCI device, we need to first
know which PCI device corresponds to each GPU Index as reported by the
FAHClient.

It appears that the GPU Index used by the FAHClient corresponds to the order
that the GPU has in the output of the FAHClient with the `--lspci` command
flag which seems to match the order of devices listed when
using the regular `lspci` command on Linux systems which is ultimately the
order of devices when sorted in ascending order by PCI device domain, bus,
slot, and then function. Using that logic, we can associate the GPU Index
in the FAHClient with a PCI bus address (domain, bus, slot, function).

The above reasoning is supported by [these findings on the Overclock.net forum](http://www.overclock.net/t/1490720/guide-configuring-client-v7-7-4-4-for-multiple-additional-gpus).

#### Method 1

##### Linux

In this method, PCI device information is determined by reading
the appropriate resources located within `/sys/bus/pci/`. (See
https://www.kernel.org/doc/Documentation/filesystems/sysfs-pci.txt for
some details about the sysfs device directory layout.)

Graphics devices are identified by reading the `class` file in each device
directory located at `/sys/bus/pci/devices/` and collecting the PCI bus
addresses for each device with a class matching `0x0300__` (the "display
controller" PCI device class)<sup>http://wiki.xomb.org/index.php?title=PCI_Class_Codes</sup>.

Then with the list of PCI bus addresses in ascending order, we can map
them to the GPU index that is reported by the FAHClient.

Now knowing which PCI devices are used by the FAHClient, we can query only
those devices for the remaining PCI information that we need. In the device's
sysfs directory (`/sys/bus/pci/devices/PCI_BUS_ADDRESS/`) we obtain the
following information: (all file contents are of the format `0x____`)

- Vendor ID: read the `vendor` file
- Device ID: read the `device` file
- Subsystem Vendor ID: read the `subsystem_vendor` file
- Subsystem Device ID: read the `subsystem_device` file

#### Method 2

##### Linux

In this method PCI device information is determined by parsing the output
of the `lspci -n -vmm -D` command and relies on the host machine having
the parent `pciutils` package already installed.

Each device is separated by two line breaks `\n\n`. Each info line has the
format of `INFO_TYPE:\tINFO_DATA\n`.

The following relevant info types and the data they represent are listed below:
- Slot: full PCI bus address
- Class, Vendor, Device, SVendor, SDevice: hex values without the preceding `0x`

Again, the output is sorted by the PCI bus address and can be linked to the
GPU Index in the FAHClient after ignoring any devices which do not have
a Class data value of `0300`.

### Obtaining GPU Device Driver Information

#### Method 1

##### Linux

Determining the GPU device driver is done by reading the file at
`/sys/bus/pci/drivers/VENDOR_NAME/module/version`.

#### Method 2

##### Linux

The appropriate vendor's driver library is employed to query the driver
itself for its version.
