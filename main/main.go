//
//Copyright (c) 2018     atighineanu
//      atighineanu@suse.de
//
//      Mr. Job-triggerer,
//       or call me... qamkins!  B-)
//##################################

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"qamkins/utils"
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
var Test_pkg_list = []string{"wicked", "glib2", "cloud-init"} //"helm", "kubernetes-salt", "sles12-velum-image", "velum"} //, "python", "yast2-hana-update", "sapconf", "haproxy"}
var machines = make(map[string]bool)
var Workdir = "/home/atighineanu/"

//-----END OF VARIABLES LIST---------

var USER = utils.SSHInfo{
	User: "atighineanu",
	Pass: "mypass",
	IP:   "my osc-having-machine IP"}

func main() {
	//----------------------CHECKING ALL THE ACTIVE UPDATES---------------------------------------
	List := utils.IssueSliceHandler()

	var a utils.PipelineCfg
	f, err := os.Open("../PipelineCfg.json")
	defer f.Close()
	if err != nil {
		log.Printf("Error: %s\n", err)
	}
	if err := json.NewDecoder(f).Decode(&a); err != nil {
		log.Printf("Error: %s\n", err)
	}

	for _, k := range Test_pkg_list {
		a.PackageName = k
		Repos, Incident := utils.FindInApi(List, k)
		if len(Repos) > 0 {
			fmt.Printf("%s \n %v\n", Repos, Incident.Base.ID)
			//fmt.Printf("%s:%v:%v %v\n", k, Incident.Base.ID, Incident.Requests.Maintenance_release[0], Incident.Checkers.Checks.Binary[0].Version)
			if Incident != nil {
				//--------------------FIRING A CONCOURSE PIPELINE--------------------------------
				job := utils.ConcourseRunner(Repos[0], *Incident, a)
				out, err := job.CombinedOutput()
				if err != nil {
					fmt.Fprintf(os.Stdout, "error: %s", err)
				}
				fmt.Println(fmt.Sprintf("%s", string(out)))
			}
		}
	}
}
