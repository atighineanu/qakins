package utils

import (
	"encoding/json"
	"flag"
	"fmt" //incident/jsonrdr/READhttp
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func HTTPREADER(url string) []byte {
	var a []byte
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			os.Exit(1)
		}
		a = contents
	}
	return a
}

// Incident represents a maintenance incident
/*type Incidents struct {
	Incident []*Incidents
} */

// Incident represents a maintenance incident
type Incident struct {
	Base       *Base
	Checkers   *Checkers
	Comments   []*Comments
	Contents   *Contents
	History    []*History
	Patchinfos *Patchinfos
	Requests   *Requests
	Smash      *Smash
	Update     *Update
}

//type history
type History struct {
	Rev    string
	Srcmd5 string
	Time   string
	User   string
}

//comments about the incident/RR
type Comments struct {
	Id   string
	Text string
	When string
	Who  string
}

//contents about maint
type Contents struct {
	Channels   []string
	Packages   []string
	Patchinfos []string
	Sources    []string
}

type Patchinfo struct {
	Category    string
	Description string
	Flags       []string
	Incident    string
	Name        string
	Packager    string
	Rating      string
	Summary     string
}

type Patchinfos struct {
	Patchinfo *Patchinfo
}

type Requests struct {
	Maintenance_incident []string
	Maintenance_release  []string
}

type Smash struct {
	Update []string
}

type Update struct {
	Crd      string
	Epoch    string
	Kpis     []string
	Prd      string
	Ratings  []string
	Severity string
}

// Base contains product data
type Base struct {
	Bugowners    []string
	Channels     []string
	Codestreams  []string
	ID           string
	Products     []string
	Project      string
	Repositories []string
	State        string
}

// Checkers are obs checks run by checker scripts
type Checkers struct {
	Checks *Checks
}

// Checks is obs checks
type Checks struct {
	Binary  []*Binary
	Install []*Install
}

// Binary is the list of archs for which the package was built
type Binary struct {
	Architecture string
	Name         string
	Command      string
	Err          string `json:"error"`
	Output       string
	Version      string
}

type Install struct {
	Architecture string
	Command      string
	Error        string
	Name         string
	Output       string
	Version      string
}

var inc = flag.String("inc", "incident", "Provide the maintenance incident")
var dl = flag.String("dl", "deadline", "Depicts the deadline field ")
var version = flag.String("v", "version", "Shows field containing the Version")
var prod = flag.String("p", "product", "Shows field for the Product name of maint.incident")
var repo = flag.String("repo", "repositories", "Shows repository list for the incident")
var arch = flag.String("arch", "", "Shows the architecture(s)/or provide arch to grep for info relevant against the arch")
var ch = flag.String("ch", "chan", "Shows the update channels")
var di = flag.String("di", "distro", "Provide the distro to search for relevant information")
var n = flag.String("n", "name", "Shows the package name for the given incident")
var src = flag.String("src", "source", "Shows the source name with version-release")

func UrlFetcher(link string) ([]byte, error) {
	response, err := http.Get(link)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 200 {
		log.Fatal("Unexpected status code", response.StatusCode)
	}
	return body, err
}

func ReadApi() []string {
	var IncidentNumberList []string
	body, err := UrlFetcher("https://maintenance.suse.de/api/incident/")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, &IncidentNumberList)
	return IncidentNumberList
}

func FindInApi(IncidentNumberList []string, Package string) string {
	var incident Incident
	var Repo string
	for i := len(IncidentNumberList) - 1000; i < len(IncidentNumberList); i++ {
		body, err := UrlFetcher("https://maintenance.suse.de/api/incident/" + IncidentNumberList[i])
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(body, &incident)
		if incident.Base.State == "active" {
			for _, k := range incident.Contents.Packages {
				if strings.Contains(k, Package) {
					//fmt.Printf("Found an update: %s  %s  %s\n", k, IncidentNumberList[i], incident.Base.Repositories)
					for _, Repositories := range incident.Base.Repositories {
						Repo = "http://download.suse.de/ibs/SUSE:/Maintenance:/" + IncidentNumberList[i] + "/" + Repositories + "/SUSE:Maintenance:" + IncidentNumberList[i] + ".repo"
						out, err := exec.Command("curl", []string{"-s", Repo}...).CombinedOutput()
						if err != nil {
							fmt.Fprintf(os.Stdout, "Couldn't open the link...%s\n", err)
						}
						tmp := fmt.Sprintf("%s", string(out))
						if strings.Contains(tmp, "key") && strings.Contains(tmp, IncidentNumberList[i]) {
							fmt.Printf("Repo for the package %s exists! Success.   ", Package)
							return Repo
						}
					}
				}
			}
		}
	}
	if Repo == "" {
		fmt.Println("Could not find any active update with this package...")
	}
	return Repo
}
