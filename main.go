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

//-----VARIABLES---------------------
var arch = "x86_64"
var Test_pkg_list = []string{"drbd", "saptune", "sapconf", "SAPHanaSR"}

//tmp := []string{"pwd"}
//basher.Bash(tmp)*/
var versions = "11|sp4 12|sp1 12|sp2 12|sp3 15| "

//-----END OF VARIABLES LIST---------

type Stamp struct {
	Time  []int64
	Month string
	Day   int64
}

type Upd struct {
	name     string
	channels []string
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

func Upd_finder() map[string]Upd {
	var Updlist, incid_info []string
	var found map[string]Upd
	found = make(map[string]Upd)
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
				incid_info = strings.Split(Updlist[i-1], ":")
				found[incid_info[3]] = Upd{name: j, channels: []string{""}}
				//tmp2.name = j
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return found
}

func Upd_chan_sorter(found map[string]Upd) map[string]incidsearch.Incident {
	var catalogue map[string]incidsearch.Incident
	for key, value := range found {
		channel := incidsearch.Incidsrc(key)
	}
	return found
}

/*
func Repo_detective(chanlist map[string][]string) {
	for _, j := range chanlist {
		for i := 0; i < len(j); i++ {
			//chanlist[j[i]] = strings.Replace(chanlist[j[i]], ":", "_", -1)
			fmt.Println(j[i])
		}
	}
}
*/

func main() {
	//fmt.Println("wohooo")
	//basher.Bash([]string{"ls", "-alh"})
	Upd_list_saver()
	//var inc = flag.String("inc,", "inci", "inci")
	//fmt.Println(Upd_finder())
	temp := Upd_chan_sorter(Upd_finder())
	fmt.Println("\n", temp)
	//Repo_detective(temp)
}
