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
	"strconv"
	"strings"
	"time"
)

//-----VARIABLES---------------------,
var arch = "x86_64"
var Test_pkg_list = []string{"drbd", "saptune", "sapconf", "SAPHanaSR"}
var machines = make(map[string]bool)

//tmp := []string{"pwd"}
//basher.Bash(tmp)
type Distri struct {
	Name    []string
	Version []string
	Flavor  []string
}

//-----END OF VARIABLES LIST---------

type Stamp struct {
	Time  []int64
	Month string
	Day   int64
}

type Upd struct {
	Inc  string
	Name []string
	Chan map[string]string
}

func Upd_list_saver() {
	var stamp Stamp
	var temp string
	//var Tmp1 []string
	var months map[string]int
	months = make(map[string]int)
	months["Jan"] = 1

	temp = basher.Bash([]string{"ls", "-alh"}, "s")
	if strings.Contains(temp, "mainlist") {
		//fmt.Printf("it is here! %s\n", temp)
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
		k, err := strconv.ParseInt(s[len(s)-4], 10, 64)
		if err != nil {
			log.Fatalf("conversion str -> int did not work! %s", err)
		}
		stamp.Day = k
		//fmt.Println(stamp)
		//fmt.Printf("%v", time.Now())
		now := time.Now()
		//fmt.Printf("%v\n%v\n%v\n%v\n", now.Day(), now.Hour(), now.Minute(), now.Month())
		a := fmt.Sprintf("%s", now.Month())
		if strings.Contains(a, stamp.Month) {
			//fmt.Println("success!!!")
			if now.Day()-int(stamp.Day) == 0 {
				if now.Hour()-int(stamp.Time[0]) > 600 {
					basher.Bash([]string{"osc", "qam", "list", ">", "mainlist"}, "p")
				}
			} else {
				basher.Bash([]string{"osc", "qam", "list", ">", "mainlist"}, "p")
			}
		} else {
			//fmt.Printf("%s %s\n", now.Month(), stamp.Month)
			fmt.Println("Oh boy... you haven't tested in a while! :-)")
			basher.Bash([]string{"osc", "qam", "list", ">", "mainlist"}, "p")
		}
	} else {
		fmt.Println("not here")
		basher.Bash([]string{"osc", "qam", "list", ">", "mainlist"}, "p")
	}
}

func Upd_finder() []Upd {
	var c int
	var Koka []Upd
	var Updlist, incid_info []string
	var found Upd

	osclist, err := os.Open("mainlist")
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
				Koka = append(Koka, found)
				Koka[c].Chan = make(map[string]string)
				incid_info = strings.Split(Updlist[i-1], ":")
				channel := incidsearch.Incidsrc(incid_info[3])
				Koka[c].Inc = incid_info[3]
				for _, l := range channel.Contents.Packages {
					Koka[c].Name = append(Koka[c].Name, l)
				}
				for _, k := range channel.Base.Channels {
					Koka[c].Chan[k] = ""
				}
				c++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return Koka
}

func Job_Distributor(Koka []Upd) {
	for i := 0; i < len(Koka); i++ {
		for key, _ := range Koka[i].Chan {
			temp := strings.Split(key, ":")
			for key2, value2 := range machines {
				temp2 := strings.Split(key2, "_")
				for k := 0; k < len(temp2); k++ {
					if strings.Contains(temp[2], temp2[2]) {
						if strings.Contains(temp[3], temp2[1]) {
							if value2 == true {
								fmt.Printf("Hooray! found machine!!! %s, for %s\n", key2, Koka[i].Name)
								value2 = false
								Koka[i].Chan[key] = key2
								machines[key2] = false
							}
						}
					}
				}
			}
		}
	}
}

func main() {
	machines["sles_11-SP4_SAP"] = true
	machines["sles_12-SP0_SAP"] = true
	machines["sles_12-SP1_SAP"] = true
	machines["sles_12-SP2_SAP"] = true
	machines["sles_12-SP3_SAP"] = true
	machines["sles_15-SP0_SAP"] = true

	Upd_list_saver()
	mp := Upd_finder()

	for _, i := range mp {
		fmt.Printf("%v\n", i)
	}
	Job_Distributor(mp)

	fmt.Printf("\n\nFollowing channgels were'nt covered yet:\n")
	for i := 0; i < len(mp); i++ {
		for key, _ := range mp[i].Chan {
			if mp[i].Chan[key] == "" {
				fmt.Printf("%s\n", key)
			}
		}
	}

	/*for key := range mp {
		for key2, value2 := range mp[key].Chan {
			fmt.Printf("BIG KEY: %v\n KEY: %v  VALUE: %v\n", key, key2, value2)
		}
	}*/

}
