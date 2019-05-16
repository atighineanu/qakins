//
//Copyright (c) 2018     atighineanu
//      atighineanu@suse.de
//
//      Mr. Job-triggerer,
//       or call me... qamkins!  B-)
//##################################

package main

import (
	"basher"
	"bufio"
	"flag"
	"fmt"
	"incidsearch"
	"log"
	"os"
	"os/exec"
	"qamkins/utils"
	"strings"
	"time"
)

const (
	usage = `You can run this program on your working machine or on a remote workstation;`
)

//-----VARIABLES---------------------
var (
	arch   = "x86_64"
	remote = flag.Bool("rem", false, "a flag that marks you run the script remote")
)

//var Test_pkg_list = []string{"drbd", "saptune", "sapconf", "SAPHanaSR", "yast2-network", "libqca2", "pacemaker"}
var Test_pkg_list = []string{"helm", "kubernetes-salt", "python", "yast2-hana-update"}
var machines = make(map[string]bool)
var Workdir = "/home/atighineanu/"

//-----END OF VARIABLES LIST---------

var USER = utils.SSHInfo{
	User: "atighineanu",
	Pass: "mypass",
	IP:   "my osc-having-machine IP"}

func Upd_list_saver() {
	flag.Parse()
	out, err := exec.Command("ls", []string{"-alh", Workdir}...).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stdout, "Couldn't execute the ls -alh $Workdir command; check your variables or Env... %s", err)
	}
	osccmd := []string{"qam", "list"}
	//osccmd := "qam list > mainlist"
	if strings.Contains(fmt.Sprintf("%s", string(out)), "mainlist") {
		now := time.Now()
		information, err := os.Lstat(Workdir + "mainlist")
		if err != nil {
			fmt.Fprintf(os.Stdout, "bad! %s", err)
		}
		createdat := information.ModTime()
		lifespan := now.Sub(createdat)
		if lifespan > 4*time.Hour {
			fmt.Println("You need to update your mainlist")
			out, err := exec.Command("osc", osccmd...).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stdout, "Couldn't osc qam command; check your variables or Env... %s", err)
			}
			file, err := os.OpenFile(Workdir+"mainlist", os.O_RDWR, os.ModeAppend)
			if err != nil {
				fmt.Fprintf(os.Stdout, "Couldn't open the mainlist file; check your variables or Env... %s", err)
			}
			len, err := file.Write(out)
			if err != nil {
				fmt.Fprintf(os.Stdout, "This amount of %v bytes failed to be copied... %s\n", len, err)
			}
			defer file.Close()
		}
	}
}

func UpdFinder() []utils.Upd {
	var c int
	var IncidJob []utils.Upd
	var Updlist, incid_info []string
	var found utils.Upd

	/*
		_, err := exec.Command("scp", USER.User+"@"+USER.IP+":/home/atighineanu/mainlist", Workdir).CombinedOutput()
		if err != nil {
			fmt.Printf("Bad! scp didn't work at copying mainlist from workstation...%s", err)
		} */

	osclist, err := os.Open(Workdir + "mainlist")
	if err != nil {
		log.Fatalf("couldn't open file...%s", err)
	}
	scanner := bufio.NewScanner(osclist)
	for scanner.Scan() {
		Updlist = append(Updlist, fmt.Sprintln(scanner.Text()))
	}

	for i := 0; i < len(Updlist); i++ {
		for _, j := range Test_pkg_list {
			if strings.Contains(Updlist[i], j) {
				fmt.Printf("Found an available update: %s\n", j)
				IncidJob = append(IncidJob, found)
				IncidJob[c].Chan = make(map[string]string)
				incid_info = strings.Split(Updlist[i-1], ":")
				channel := incidsearch.Incidsrc(incid_info[3])
				IncidJob[c].Inc = incid_info[3]
				for _, l := range channel.Contents.Packages {
					IncidJob[c].Name = append(IncidJob[c].Name, l)
				}
				for _, k := range channel.Base.Channels {
					IncidJob[c].Chan[k] = ""
				}
				c++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return IncidJob
}

func RepoHandler(cov []utils.Upd) {
	const prefixurl = "http://download.suse.de/ibs/SUSE:/Maintenance:/"
	for i := 0; i < len(cov); i++ {
		for j, _ := range cov[i].Chan {
			if cov[i].Chan[j] != "WAITING" && cov[i].Chan[j] != "FAIL." {
				reposufix := "/SUSE:Maintenance:" + cov[i].Inc + ".repo"
				updlink := prefixurl + cov[i].Inc + "/" + strings.Replace(j, ":", "_", -1)
				repolink := updlink + reposufix
				fmt.Println(repolink)
			}
		}
	}
}

func JobStarter(IncidJob []utils.Upd) {
	for i := 0; i < len(IncidJob); i++ {
		for key, _ := range IncidJob[i].Chan {
			if strings.Contains(IncidJob[i].Chan[key], ":") {
				basher.Bash([]string{"sudo", "virsh", "domifaddr", IncidJob[i].Chan[key], "--source", "agent", "|", "grep", "eth", "|", "awk '{print $4}'"}, "s")
			}
		}
	}
}

func JobDistributor(IncidJob []utils.Upd) []utils.Upd {
	flg := false
	var c int                            //--counter to count the number of machines we checked
	for i := 0; i < len(IncidJob); i++ { //----for how many update incidents were found
		for key, _ := range IncidJob[i].Chan { //----for how many channels per current update
			temp := strings.Split(key, ":") //---temp is an array of strings with, for example temp[0]    [1]      [2]	      [3]           [4]
			// ":" - is the symbol of splitting	                                      			   SUSE:Updates:SLE-SERVER:11-SP1-TERADATA:x86_64
			// machines - is a map (hashtable in go) where key is the machine name, while the bool value is true or false (depending is the machine free or busy)
			// - the reason of this is that there might be updates with pretty much overlapping to be tested on the same VMs
			for key2, value2 := range machines {
				temp2 := strings.Split(key2, ":") //---its the same kind of array like temp above
				//--- now it searches through "SLE-HA", or "SLE-SERVER" (field #2), then field "11-SP4" or "12-SP1" (field #3 of temp and #1 of temp2), and so on...
				if strings.Contains(temp[2], temp2[2]) && strings.Contains(temp[3], temp2[1]) && strings.Contains(temp[4], temp2[3]) {
					c = 0
					flg = true
					if value2 == true {
						fmt.Printf("Found machine!!! %s, for %s %s\n", key2, IncidJob[i].Name, IncidJob[i].Inc)
						value2 = false
						IncidJob[i].Chan[key] = key2
						machines[key2] = false //---- setting "FREE" at the machine key to "false" -> machine is busy
					} else {
						c = 0
						flg = true
						IncidJob[i].Chan[key] = "WAITING" //---it means a machine for this specific channel exists, but it is busy
					}
				} else {
					c++
					flg = false
				}
				if c == len(machines) && flg == false {
					IncidJob[i].Chan[key] = "FAIL." //--- it means it didn't find any VM with channel's parameters
					c = 0
				}
			}
		}
	}
	return IncidJob
}

func main() {

	//		PROD  DISTRO FLAV  ARCH
	machines["SLE:11-SP4:HA:x86_64"] = true
	machines["SLE:12-SP0:HA:x86_64"] = true
	machines["SLE:12-SP1:HA:x86_64"] = true
	machines["SLE:12-SP2:HA:x86_64"] = true
	machines["SLE:12-SP3:HA:x86_64"] = true
	machines["SLE:15-SP0:HA:x86_64"] = true
	machines["SLE:11-SP4:SAP:x86_64"] = true
	machines["SLE:12-SP0:SAP:x86_64"] = true
	machines["SLE:12-SP1:SAP:x86_64"] = true
	machines["SLE:12-SP2:SAP:x86_64"] = true
	machines["SLE:12-SP3:SAP:x86_64"] = true
	machines["SLE:15-SP0:SAP:x86_64"] = true
	machines["SLE:11-SP4:SERVER:x86_64"] = true

	/*
		Upd_list_saver()
		mp := UpdFinder()
		//var cov []utils.Upd
		if len(mp) == 0 {
			fmt.Println("That's bad!")
		} else {
			fmt.Printf("These Updates where found:\n")
			for i := 0; i < len(mp); i++ {
				for key, _ := range mp[i].Chan {
					fmt.Printf("Incident: %s   Name: %s   Channel: %s\n", mp[i].Inc, mp[i].Name, key)
				}
			}
			//cov = JobDistributor(mp)
			fmt.Printf("Following channgels were'nt covered yet:\n")
			for i := 0; i < len(mp); i++ {
				for key, _ := range mp[i].Chan {
					if mp[i].Chan[key] == "FAIL." {
						fmt.Printf("%s  %s  %s\n", mp[i].Inc, mp[i].Name, key)
					}
				}
			}
		}

		RepoHandler(mp)
		for i := 0; i < len(cov); i++ {
			fmt.Printf("\n%v - %v\n", cov[i].Inc, cov[i].Name)
			for key, value := range cov[i].Chan {
				fmt.Printf("%v  -  %v\n", key, value)
			}
		}*/

	List := utils.ReadApi()

	for _, k := range Test_pkg_list {
		Repos := utils.FindInApi(List, k)
		if len(Repos) > 0 {
			fmt.Printf("%s: %s\n", k, Repos)
		}
	}
}
