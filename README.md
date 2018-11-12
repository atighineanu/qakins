qamkins - automation tool for on-spot 
testing qam@SUSE updates.

#H3 Requirements:

#H4 - qemu-kvm, libvirt installed

#H4 - VMs:
	* should have installed qemu-guest-agent
	  (properly setup & running)

	* should have a specific syntax of <domname>
	  SLE:11-SP4:HA:x86_64
	   1    2     3    4
		   1    2      3   4

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


 
