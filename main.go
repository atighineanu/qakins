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
	"fmt"
	"incidsearch"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

//-----VARIABLES---------------------
var arch = "x86_64"

//var Test_pkg_list = []string{"drbd", "saptune", "sapconf", "SAPHanaSR", "yast2-network", "libqca2", "pacemaker"}
var Test_pkg_list = []string{"helm", "kubernetes-salt", "python", "gnutls"}
var machines = make(map[string]bool)
var Workdir = "/Users/alexeitighineanu/automation"

//-----END OF VARIABLES LIST---------

type Stamp struct {
	Time  []int64
	Month string
	Day   int64
}

/* may be obsolete, let's see...
type Job struct {
	Machine  string
	Repolink string
}
*/

var USER = SSHInfo{
	User: "atighineanu",
	Pass: "mypass",
	IP:   "my osc-having-machine IP"}

type Upd struct {
	Inc  string
	Name []string
	Chan map[string]string
}

type SSHInfo struct {
	User string
	Pass string
	IP   string
}

func (s *SSHInfo) Command(cmd ...string) *exec.Cmd {

	arg := append(
		[]string{"-o", "StrictHostKeyChecking=no",
			fmt.Sprintf("%s@%s", s.User, s.IP),
		},
		cmd...,
	)
	return exec.Command("ssh", arg...)
}

func SSHPrinter(command []string, user SSHInfo) {

	out, err := user.Command(command...).CombinedOutput()
	if err != nil {
		fmt.Printf("Bad! Error at runng osc thourgh SSH... %s\n", err)
	}
	fmt.Println(fmt.Sprintf("%s", string(out)))
}

func Upd_list_saver() {
	var stamp Stamp
	var temp string
	var months map[string]int
	months = make(map[string]int)
	months["Jan"] = 1

	temp = basher.Bash([]string{"ls", "-alh", Workdir}, "s")
	osccmd := []string{"osc", "qam", "list", ">", "mainlist"}
	if strings.Contains(temp, "mainlist") {
		s := strings.Split(temp, " ")
		tmp1 := strings.Split(s[len(s)-2], ":")

		for _, i := range tmp1 {
			k, err := strconv.ParseInt(i, 10, 64)
			if err != nil {
				log.Fatalf("conversion str -> int did not work! %s", err)
			}
			stamp.Time = append(stamp.Time, k)
		}
		stamp.Month = s[len(s)-3]
		///on Mac is len(s) - 3
		k, err := strconv.ParseInt(s[len(s)-3], 10, 64)
		if err != nil {
			log.Fatalf("conversion str -> int did not work! %s", err)
		}
		stamp.Day = k
		now := time.Now()
		//fmt.Printf("%v\n%v\n%v\n%v\n", now.Day(), now.Hour(), now.Minute(), now.Month())
		a := fmt.Sprintf("%s", now.Month())
		if strings.Contains(a, stamp.Month) {
			if now.Day()-int(stamp.Day) == 0 {
				if now.Hour()-int(stamp.Time[0]) > 600 {
					SSHPrinter(osccmd, USER)
				}
			} else {
				SSHPrinter(osccmd, USER)
			}
		} else {
			SSHPrinter(osccmd, USER)
		}
	} else {
		fmt.Println("not here")
		SSHPrinter(osccmd, USER)
	}
}

func UpdFinder() []Upd {
	var c int
	var IncidJob []Upd
	var Updlist, incid_info []string
	var found Upd

	_, err := exec.Command("scp", USER.User+"@"+USER.IP+":/home/atighineanu/mainlist", Workdir).CombinedOutput()
	if err != nil {
		fmt.Printf("Bad! scp didn't work at copying mainlist from workstation...%s", err)
	}

	osclist, err := os.Open(Workdir + "/mainlist")
	if err != nil {
		log.Fatalf("couldn't open file...%s", err)
	}
	scanner := bufio.NewScanner(osclist)
	for scanner.Scan() {
		Updlist = append(Updlist, fmt.Sprintln(scanner.Text()))
	}

	//fmt.Printf("%v", Updlist)

	for i := 0; i < len(Updlist); i++ {
		for _, j := range Test_pkg_list {
			if strings.Contains(Updlist[i], j) {
				fmt.Printf("yohooo! %s\n", j)
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

func RepoHandler(cov []Upd) {
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

func JobStarter(IncidJob []Upd) {
	for i := 0; i < len(IncidJob); i++ {
		for key, _ := range IncidJob[i].Chan {
			if strings.Contains(IncidJob[i].Chan[key], ":") {
				basher.Bash([]string{"sudo", "virsh", "domifaddr", IncidJob[i].Chan[key], "--source", "agent", "|", "grep", "eth", "|", "awk '{print $4}'"}, "s")
			}
		}
	}
}

func JobDistributor(IncidJob []Upd) []Upd {
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

	Upd_list_saver()
	mp := UpdFinder()
	var cov []Upd
	if len(mp) == 0 {
		fmt.Println("rau way!")
	} else {
		fmt.Printf("These Updates where found:\n")
		for i := 0; i < len(mp); i++ {
			for key, _ := range mp[i].Chan {
				fmt.Printf("Incident: %s   Name: %s   Channel: %s\n", mp[i].Inc, mp[i].Name, key)
			}
		}
		cov = JobDistributor(mp)
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
	}
}
