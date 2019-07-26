package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/PuerkitoBio/goquery"
)

type Team struct {
	name          string
	wins          int
	losses        int
	circuitPoints int
}

type Match struct {
	blue   *Team
	red    *Team
	winner *Team
}

func (m Match) GetLoser() *Team {
	if m.winner == nil {
		return nil
	}

	if m.red == m.winner {
		return m.blue
	} else {
		return m.red
	}
}

func (m Match) String() string {
	var s string
	if m.winner == nil {
		s = ""
	} else {
		s = m.winner.name
	}
	return fmt.Sprintf("%s vs %s (w=%s)", m.blue.name, m.red.name, s)
}

type Schedule struct {
	matches []Match
}

type Rank struct {
	team *Team
	tie  bool
}

type Season struct {
	standings []Rank
	schedule  Schedule
}

func (s Season) Sort() {
	for _, r := range s.standings {
		r.tie = false
	}

	sort.Slice(s.standings, func(i int, j int) bool {
		return s.standings[i].team.wins > s.standings[j].team.wins
	})

	for i, _ := range s.standings {
		for j, _ := range s.standings {
			if s.standings[i].team.wins == s.standings[j].team.wins &&
				s.standings[i].team.name != s.standings[j].team.name {
				s.standings[i].tie = true
				s.standings[j].tie = true
			}
		}
	}
}

func (s Season) String() string {
	str := ""
	for _, r := range s.standings {
		str += fmt.Sprintf("%s (%d-%d) (%v) (@[%p])\n", r.team.name, r.team.wins, r.team.losses, r.tie, &r.team)
	}

	return str
}

func GetSelectorString(week int) string {
	return fmt.Sprintf(".ml-allw.ml-w%d.ml-row", week)
}

func GetLCSTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["Team Liquid"] = &Team{"TL", 0, 0, 90}
	m["Team SoloMid"] = &Team{"TSM", 0, 0, 70}
	m["Cloud9"] = &Team{"C9", 0, 0, 40}
	m["FlyQuest"] = &Team{"FLY", 0, 0, 40}
	m["Echo Fox"] = &Team{"FOX", 0, 0, 10}
	m["Golden Guardians"] = &Team{"GGS", 0, 0, 10}
	m["100 Thieves"] = &Team{"100T", 0, 0, 0}
	m["Clutch Gaming"] = &Team{"CG", 0, 0, 0}
	m["Counter Logic Gaming"] = &Team{"CLG", 0, 0, 0}
	m["OpTic Gaming"] = &Team{"OPT", 0, 0, 0}
	return m
}

func ParseSchedule(url string, teams map[string]*Team) Schedule {
	teams[""] = nil
	defer delete(teams, "")

	s := Schedule{[]Match{}}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	for i := range [10]int{} {
		cls := GetSelectorString(i + 1)
		doc.Find(cls).Each(func(index int, element *goquery.Selection) {
			children := element.ChildrenFiltered(".ml-team")
			var winner *Team
			winner = nil

			var red string
			children.Slice(1, 2).Each(func(i int, el *goquery.Selection) {
				red, _ = el.Attr("data-teamhighlight")
				if el.HasClass("matchlist-winner-team") {
					winner = teams[red]
				}
			})

			var blue string
			children.Slice(0, 1).Each(func(i int, el *goquery.Selection) {
				blue, _ = el.Attr("data-teamhighlight")
				if el.HasClass("matchlist-winner-team") {
					winner = teams[blue]
				}
			})

			m := Match{teams[blue], teams[red], winner}
			s.matches = append(s.matches, m)
			if winner != nil {
				winner.wins++
				m.GetLoser().losses++
			}
		})
	}

	return s
}
