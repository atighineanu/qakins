qamkins - automation tool for on-spot 
testing qam@SUSE updates.

#H3 Requirements:

#H4 - qemu-kvm, libvirt installed

#H4 - VMs:
	
	* should have installed qemu-guest-agent
	  (properly setup & running)
	 Note: for old distros qemu-ga does not run
	 after installation and addig another Virtio 
	 controller to the VM.
	 Try to virsh edit <VM>, and add next to the 
	 "SPICE" channel in the XML, which looks like:
	 <channel type='spicevmc'>
           <target type='virtio' name='com.redhat.spice.0'/>
           <address type='virtio-serial' controller='0' bus='0' port='1'/>
          </channel>

	 add this:
	 <channel type='unix'>
      	   <target type='virtio' name='org.qemu.guest_agent.0'/>
      	   <address type='virtio-serial' controller='0' bus='0' port='2'/>
         </channel>
	  (be aware that if you have more virtio cont-
	 rollers, you don't mix the port numbers, and
	 in your case it might be port='3' or '4' and
	 so on...)
	 
	 restart qemu-ga (rcqemu-ga for 11sp4)


	* should have a specific syntax of <domname>
	  SLE:11-SP4:HA:x86_64
	   1    2     3    4
	  ===================
	      ,where:
		1 - Product
		2 - Version
		3 - Flavour (you can pile up more
		modules like HA_SDK_BASE_SAP_WE)
		4 - Architecture

	   !Note: It is recommended to add stuff 
	    AFTER the section "4", e.g. - a suffix 
	    to the VM name (whatever you want).
	    If you want to add a prefix - do not use
	    ":" before section "1" (you can use any-
	    thing else "_", "-", e.t.c.)

	* should have your SSH keys installed, fire-
	  wall settings for port 22 covered, and 
	  obviously sshd running. (beware: on 
	  openSUSE LEAP sshd.service sometimes stops!)

	* should have connection to network, or...
	  if they are isolated -> should have locally
	  mounted update repositories. 
 
