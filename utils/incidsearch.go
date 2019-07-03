package utils

import (
	"encoding/json"
	"fmt" //incident/jsonrdr/READhttp
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
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

func IssueSliceHandler() []string {
	var f *os.File
	t := time.Now()
	fileinfo, err := os.Lstat("IssueNumberList")
	if err != nil {
		fmt.Fprintf(os.Stdout, "IssueSliceSaver->os.Lstat: file cannot be found!  %s\n", err)
		f, err = os.Create("IssueNumberList")
		if err != nil {
			fmt.Fprintf(os.Stdout, "IssueSliceSaver->os.Create: file couldn't be created!  %s\n", err)
		}
		buf, err := json.Marshal(ReadApi())
		if err != nil {
			fmt.Fprintf(os.Stdout, "IssueSliceSaver->json.Marshal: couldn't Marshal data!  %s\n", err)
		}
		_, err = f.Write(buf) //I don't need (so far) to know the exact amount of bytes written to file
		if err != nil {
			fmt.Fprintf(os.Stdout, "IssueSliceSaver->f.Write: couldn't write to file!  %s\n", err)
		}
		defer f.Close()
	} else {
		created := fileinfo.ModTime()
		if t.Sub(created) > 24*time.Hour {
			f, err = os.OpenFile("IssueNumberList", os.O_RDWR, 0644)
			if err != nil {
				fmt.Fprintf(os.Stdout, "IssueSliceSaver->os.OpenFile: file couldn't be opened!  %s\n", err)
			}
			buf, err := json.Marshal(ReadApi())
			if err != nil {
				fmt.Fprintf(os.Stdout, "IssueSliceSaver->json.Marshal: couldn't Marshal data!  %s\n", err)
			}
			_, err = f.Write(buf) //I don't need (so far) to know the exact amount of bytes written to file
			if err != nil {
				fmt.Fprintf(os.Stdout, "IssueSliceSaver->f.Write: couldn't write to file!  %s\n", err)
			}
			defer f.Close()
		}
	}
	f, err = os.OpenFile("IssueNumberList", os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintf(os.Stdout, "IssueSliceSaver->os.OpenFile: file couldn't be opened!  %s\n", err)
	}

	var content []byte
	var IncidentNumberList []string
	content, err = ioutil.ReadFile("IssueNumberList")
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stdout, "IssueSliceSaver->f.Read: file couldn't be red!  %s\n", err)
	}
	json.Unmarshal(content, &IncidentNumberList)
	return IncidentNumberList
}

func FindInApi(IncidentNumberList []string, Package string) ([]string, *Incident) {
	var incident, found Incident
	var Repos []string
	for i := 1; /*len(IncidentNumberList) - 1000*/ i < len(IncidentNumberList); i++ { // HERE I SWEEP TRHOUGH ALL ISSUES 1 - 12K
		body, err := UrlFetcher("https://maintenance.suse.de/api/incident/" + IncidentNumberList[i])
		if err != nil {
			log.Fatal(err)
		}
		json.Unmarshal(body, &incident) //UNMARSHALLING THE RESPONSE OF MAINTENANCE API INTO AN "INCIDENT" DATA TYPE
		if incident.Base.State == "active" {
			for _, k := range incident.Contents.Packages {
				if strings.Contains(k, Package) {
					for _, Repository := range incident.Base.Repositories {
						if strings.Contains(Repository, "x86_64") && !strings.Contains(Repository, "DEBUG") { //EXCLUDING DEBUG, COVERING x86_64 (SO FAR)
							// THE REPOSITORY STRUCTURE IS PRETTY HARDCODED...
							Repo := "http://download.suse.de/ibs/SUSE:/Maintenance:/" + IncidentNumberList[i] + "/" + Repository + "/SUSE:Maintenance:" + IncidentNumberList[i] + ".repo"
							out, err := exec.Command("curl", []string{"-s", Repo}...).CombinedOutput()
							if err != nil {
								fmt.Fprintf(os.Stdout, "Couldn't open the link...%s\n", err)
							}
							tmp := fmt.Sprintf("%s", string(out))

							// CHECKING IF .repo FILE FROM COMPOSED LINK IS REAL
							if strings.Contains(tmp, "key") && strings.Contains(tmp, IncidentNumberList[i]) {
								log.Printf("Repo for the package %s exists! Success.   \n", Package)
								count := 0
								// HERE I CHECK FOR DUPLICATE .Repo FILES... AND THEN ADDING IT TO Repos SLICE
								for _, k := range Repos {
									if k == Repo {
										count++
									}
								}
								if count == 0 && strings.Contains(strings.ToLower(Repo), "caasp") && strings.Contains(Repo, "3.0") {
									Repos = append(Repos, Repo)
								}
								json.Unmarshal(body, &found)
							}
						}
					}
				}
			}
		}
		time.Sleep(1000 * time.Microsecond)
	}
	if len(Repos) == 0 {
		fmt.Println("Could not find any active update with this package...")
	}
	return Repos, &found
}
