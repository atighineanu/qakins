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
	USER = utils.SSHInfo{
		User: "atighineanu",
		Pass: "mypass",
		IP:   "my osc-having-machine IP"}
	//----------------------FLAGS------------------------------------
	howto      = flag.Bool("help", false, "prints how to use instructions")
	dockuser   = flag.String("dockuser", "", "docker username")
	dockpasswd = flag.String("dockpasswd", "", "docker password")
	dockrepo   = flag.String("dockrepo", "", "your docker repository")
	pipename   = flag.String("pipename", "", "your concourse pipeline name")
	packname   = flag.String("packname", "", "the package name")
	justsearch = flag.Bool("justsearch", false, "just searching through Maint API")
	login      = flag.Bool("login", false, "triggers login of fly")
	triggerjob = flag.Bool("triggerjob", false, "triggers a job")

	//----------------------PACKAGES-TO-TEST-LIST---------------------
	//var Test_pkg_list = []string{"drbd", "saptune", "sapconf", "SAPHanaSR", "yast2-network", "libqca2", "pacemaker"}
	Test_pkg_list = []string{"kubernetes-salt", "cloud-init", "velum", "python-MarkupSafe", "zeromq", "velum", "wicked", "glib2"} //"helm", "kubernetes-salt", "sles12-velum-image", "velum"} //, "python", "yast2-hana-update", "sapconf", "haproxy"}
	howtoconst    = `1. Make sure you fill proper information in /qakins/PipelineCfg.json
	2. Make sure you indicate proper package names when you search for updates
	   you can run: # main(or #go run main.go) -packname apache2 -justsearch
	   in case you want to spawn concourse pipelines - you just run: # main 
		- this way, the pipeline will be named <<PackageName:ProjectName:UpdateSeverity>>
	   if you want to run a pipeline with custom settings, then flags like -pipename 
	   might be handy`
)

//-----END OF VARIABLES LIST---------

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
	//----------------------CHECKING IF SOME PARAMS WERE PASSED FROM COMMANDLINE-------------------
	flag.Parse()
	var Testlist []string
	if *howto {
		fmt.Printf("%s", howtoconst)
	}
	if *dockuser != "" {
		a.Username = *dockuser
	}
	if *dockpasswd != "" {
		a.Password = *dockpasswd
	}
	if *dockrepo != "" {
		a.DockerRepo = *dockrepo
	}
	if *packname != "" {
		a.PackageName = *packname
		Testlist = append(Testlist, *packname)
	} else {
		Testlist = Test_pkg_list
	}
	if *pipename != "" {
		a.PipeName = *pipename
	}
	//-------------------------------------EXECUTION ----------------------------------------------
	if *login {
		utils.FlyLogin()
	} else {
		for _, k := range Testlist {
			a.PackageName = k
			Repos, Incident := utils.FindInApi(List, k)
			if len(Repos) > 0 {
				fmt.Printf("%s \n %v\n", Repos, Incident.Base.ID)
				//fmt.Printf("%s:%v:%v %v\n", k, Incident.Base.ID, Incident.Requests.Maintenance_release[0], Incident.Checkers.Checks.Binary[0].Version)
				if Incident != nil {
					//--------------------FIRING A CONCOURSE PIPELINE--------------------------------
					if !*justsearch {
						utils.FlyLogin()
						job, pipename := utils.ConcourseRunner(Repos[0], *Incident, a)
						out, err := job.CombinedOutput()
						if err != nil {
							fmt.Fprintf(os.Stdout, "error: %s", err)
						}
						fmt.Println(fmt.Sprintf("%s", string(out)))
						if *triggerjob {
							utils.FlyJobTrigg(pipename)
						}
					} else {
						fmt.Println(Repos)
					}
				}
			}
		}
	}
}
