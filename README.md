#qamkins - automation tool for on-spot

####testing qam@SUSE updates.

#### * What it does:
Qamkins lists all the updates in the QAM queue, if there are some matching the desired package names - it searches for available VMs (qemu-kvm, libvirt, didn't test with XEN yet) and fires a test for respective update. Qamkins processes not just the S:M:INC:RR update, but also all it's channels. If a machine corresponding to the update channel is found and free - the job starts. If the machine is found but busy - channel will be registered as "Wainting". And, if the correspondig machine for the update channel isn't found - it is FAIL.


## Requirements:

### - qemu-kvm, libvirt installed

### - VMs:

* should have installed qemu-guest-agent
 (properly setup & running)
 
     - Note: for old distros qemu-ga does not run after installation and adding another Virtio controller to the VM. Try to virsh edit <VM>, and add next to the "SPICE" channel in the XML, which looks like:
 
```
   <channel type='spicevmc'>
      <target type='virtio' name='com.redhat.spice.0'/>
	  <address type='virtio-serial' controller='0' bus='0' port='1'/>
   </channel>
```

 add this: 
 
```
 <channel type='unix'>
 <target type='virtio' name='org.qemu.guest_agent.0'/>
 <address type='virtio-serial' controller='0' bus='0' port='2'>
 </channel>
```
    - (be aware that if you have more virtio controllers, you don't mix the port numbers, and in your case it might be port='3' or '4' and so on...)
 
    - restart qemu-ga (rcqemu-ga for 11sp4)

* should have a specific syntax of <domname>

  ```
  SLE:11-SP4:HA:x86_64
     |    |  |    |
     1    2  3    4
  ```
  ,where:
  - 1 - Product
  - 2 - Version
  - 3 - Flavour (you can pile up more
	 modules like```HA_SDK_BASE_SAP_WE```)
  - 4 - Architecture

 - !Note: It is recommended to add stuff 
    AFTER the section "4", e.g. - a suffix 
    to the VM name (whatever you want).
    If you want to add a prefix - do not use
    ":" before section "1" (you can use any-
    thing else "_", "-", e.t.c.). Elements 1,2,3 and 4 should be isolated with ":" between each-other.

* should have your SSH keys installed, fire-
  wall settings for port 22 covered, and 
  obviously sshd running. (beware: on 
  openSUSE LEAP sshd.service sometimes stops!)

* should have connection to network, or...
  if they are isolated -> should have locally
  mounted update repositories.


### - osc
Set up your osc alias in .bashrc or set it for your session alias="correspondingurl"


### - sudoers
Set up your user to be in the sudoers list (e.g. passwordless sudo)




 
