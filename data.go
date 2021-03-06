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
	simmed bool
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
	return fmt.Sprintf("%s vs %s (w=%s) [sim=%v]", m.blue.name, m.red.name, s, m.simmed)
}

type Schedule struct {
	matches []*Match
}

type Rank struct {
	team *Team
	tie  bool
}

type Season struct {
	standings []Rank
	schedule  Schedule
}

func (s Season) Sort(lec bool) {
	for _, r := range s.standings {
		r.tie = false
	}

	doNotTie := []string{}

	sort.Slice(s.standings, func(i int, j int) bool {
		iTeam := s.standings[i].team
		jTeam := s.standings[j].team
		iWins := 0
		jWins := 0
		iWinsSecondHalf := 0
		jWinsSecondHalf := 0
		for z, m := range s.schedule.matches {
			if m.winner == nil {
				continue
			}
			if m.winner.name == iTeam.name && m.GetLoser().name == jTeam.name {
				iWins++
			} else if m.winner.name == jTeam.name && m.GetLoser().name == iTeam.name {
				jWins++
			}

			if m.winner.name == iTeam.name {
				if z >= 44 {
					iWinsSecondHalf++
				}
			} else if m.winner.name == jTeam.name {
				if z >= 44 {
					jWinsSecondHalf++
				}
			}
		}
		if iTeam.wins == jTeam.wins && iTeam.losses == jTeam.losses {
			if iWins != jWins {
				// fmt.Printf("Breaking h2h tie between %v and %v\n", s.standings[i].team.name, s.standings[j].team.name)
				doNotTie = append(doNotTie, fmt.Sprintf("%v|%v", iTeam.name, jTeam.name))
				doNotTie = append(doNotTie, fmt.Sprintf("%v|%v", jTeam.name, iTeam.name))
				return iWins > jWins
			} else if lec {
				if iWinsSecondHalf != jWinsSecondHalf {	
					doNotTie = append(doNotTie, fmt.Sprintf("%v|%v", iTeam.name, jTeam.name))
					doNotTie = append(doNotTie, fmt.Sprintf("%v|%v", jTeam.name, iTeam.name))
					if (iTeam.name == "G2" || iTeam.name == "FNC") && (jTeam.name == "G2" || jTeam.name == "FNC") {
						// fmt.Printf("%v vs %v in 2nd half (%v to %v)\n", iTeam.name, jTeam.name, iWinsSecondHalf, jWinsSecondHalf)
					}
				}
				
				return iWinsSecondHalf > jWinsSecondHalf
			} else {
				return iWins > jWins
			}
			// fmt.Printf("Checking h2h between %v and %v -- %v-%v\n", iTeam.name, jTeam.name, iWins, jWins)
		} else if iTeam.losses != jTeam.losses {
			return iTeam.losses < jTeam.losses
		} else {
			return iTeam.wins > jTeam.wins
		}
	})

	for i, _ := range s.standings {
		for j, _ := range s.standings {
			if s.standings[i].team.wins == s.standings[j].team.wins &&
				s.standings[i].team.name != s.standings[j].team.name &&
				s.standings[i].team.losses == s.standings[j].team.losses {
				if !contains(doNotTie, fmt.Sprintf("%v|%v", s.standings[i].team.name, s.standings[j].team.name)) {
					s.standings[i].tie = true
					s.standings[j].tie = true
				} else {
					// fmt.Printf("Breaking h2h tie between %v and %v\n", s.standings[i].team.name, s.standings[j].team.name)
				}
			}
		}
	}
}

func (s Season) String() string {
	str := ""
	for _, r := range s.standings {
		str += fmt.Sprintf("%s (%d-%d) (%v)\n", r.team.name, r.team.wins, r.team.losses, r.tie)
	}

	return str
}

func GetSelectorString(week int) string {
	return fmt.Sprintf(".ml-allw.ml-w%d.ml-row", week)
}

func MakeTeam(name string, circuitPoints int) *Team {
	return &Team{name, 0, 0, circuitPoints}
}

// func GetLCSTeams() map[string]*Team {
// 	m := make(map[string]*Team)
// 	m["Team Liquid"] = &Team{"TL", 0, 0, 90}
// 	m["Team SoloMid"] = &Team{"TSM", 0, 0, 70}
// 	m["Cloud9"] = &Team{"C9", 0, 0, 40}
// 	m["FlyQuest"] = &Team{"FLY", 0, 0, 40}
// 	m["Echo Fox"] = &Team{"FOX", 0, 0, 10}
// 	m["Golden Guardians"] = &Team{"GGS", 0, 0, 10}
// 	m["100 Thieves"] = &Team{"100T", 0, 0, 0}
// 	m["Clutch Gaming"] = &Team{"CG", 0, 0, 0}
// 	m["Counter Logic Gaming"] = &Team{"CLG", 0, 0, 0}
// 	m["OpTic Gaming"] = &Team{"OPT", 0, 0, 0}
// 	return m
// }

func GetLCSTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["Team Liquid"] = &Team{"TL", 0, 0, 0}
	m["TSM"] = &Team{"TSM", 0, 0, 0}
	m["Cloud9"] = &Team{"C9", 0, 0, 0}
	m["FlyQuest"] = &Team{"FLY", 0, 0, 0}
	m["Immortals"] = &Team{"IMT", 0, 0, 0}
	m["Golden Guardians"] = &Team{"GG", 0, 0, 0}
	m["100 Thieves"] = &Team{"100T", 0, 0, 0}
	m["Dignitas"] = &Team{"DIG", 0, 0, 0}
	m["Counter Logic Gaming"] = &Team{"CLG", 0, 0, 0}
	m["Evil Geniuses.NA"] = &Team{"EG", 0, 0, 0}
	return m
}

// func GetLECTeams() map[string]*Team {
// 	m := make(map[string]*Team)
// 	m["G2 Esports"] = MakeTeam("G2", 0)
// 	m["Fnatic"] = MakeTeam("FNC", 0)
// 	m["Splyce"] = MakeTeam("SPY", 0)
// 	m["FC Schalke 04 Esports"] = MakeTeam("S04", 0)
// 	m["Origen"] = MakeTeam("OG", 0)
// 	m["Rogue (European Team)"] = MakeTeam("RGE", 0)
// 	m["Team Vitality"] = MakeTeam("VIT", 0)
// 	m["Misfits Gaming"] = MakeTeam("MSF", 0)
// 	m["SK Gaming"] = MakeTeam("SK", 0)
// 	m["Excel Esports"] = MakeTeam("XL", 0)
// 	return m
// }

func GetLECTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["G2 Esports"] = MakeTeam("G2", 0)
	m["Fnatic"] = MakeTeam("FNC", 0)
	m["MAD Lions"] = MakeTeam("MAD", 0)
	m["FC Schalke 04 Esports"] = MakeTeam("S04", 0)
	m["Astralis"] = MakeTeam("OG", 0)
	m["Rogue (European Team)"] = MakeTeam("RGE", 0)
	m["Team Vitality"] = MakeTeam("VIT", 0)
	m["Misfits Gaming"] = MakeTeam("MSF", 0)
	m["SK Gaming"] = MakeTeam("SK", 0)
	m["Excel Esports"] = MakeTeam("XL", 0)
	return m
}

func GetLCKTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["SANDBOX Gaming"] = MakeTeam("SB", 0)
	m["DAMWON Gaming"] = MakeTeam("DWG", 0)
	m["Griffin"] = MakeTeam("GRF", 0)
	m["Kingzone DragonX"] = MakeTeam("KZ", 0)
	m["Gen.G"] = MakeTeam("GEN", 0)
	m["SK Telecom T1"] = MakeTeam("SKT", 0)
	m["Afreeca Freecs"] = MakeTeam("AF", 0)
	m["KT Rolster"] = MakeTeam("KT", 0)
	m["Hanwha Life Esports"] = MakeTeam("HLE", 0)
	m["Jin Air Green Wings"] = MakeTeam("JAG", 0)
	return m
}

func GetLPLTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["Invictus Gaming"] = MakeTeam("IG", 90)
	m["JD Gaming"] = MakeTeam("JDG", 70)
	m["FunPlus Phoenix"] = MakeTeam("FPX", 50)
	m["Top Esports"] = MakeTeam("TES", 30)
	m["Royal Never Give Up"] = MakeTeam("RNG", 10)
	m["Dominus Esports"] = MakeTeam("DMO", 10)
	m["Bilibili Gaming"] = MakeTeam("BLG", 0)
	m["EDward Gaming"] = MakeTeam("EDG", 0)
	m["LGD Gaming"] = MakeTeam("LDG", 0)
	m["Oh My God"] = MakeTeam("OMG", 0)
	m["Rogue Warriors"] = MakeTeam("RW", 0)
	m["LNG Esports"] = MakeTeam("LNG", 0)
	m["Suning"] = MakeTeam("SN", 0)
	m["Team WE"] = MakeTeam("WE", 0)
	m["Vici Gaming"] = MakeTeam("VG", 0)
	m["Victory Five"] = MakeTeam("V5", 0)
	return m
}

func GetLMSTeams() map[string]*Team {
	m := make(map[string]*Team)
	m["J Team"] = MakeTeam("JT", 30)
	m["ahq e-Sports Club"] = MakeTeam("AHQ", 50)
	m["MAD Team"] = MakeTeam("MAD", 70)
	m["Flash Wolves"] = MakeTeam("FW", 90)
	m["Hong Kong Attitude"] = MakeTeam("HKA", 10)
	m["G-Rex"] = MakeTeam("GRX", 10)
	m["Alpha Esports"] = MakeTeam("ALF", 0)
	return m
}

func ParseSchedule(url string, teams map[string]*Team) Schedule {
	teams[""] = nil
	defer delete(teams, "")

	s := Schedule{[]*Match{}}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	for i := range [12]int{} {
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
			if red == "" {
				return
			}

			var blue string
			children.Slice(0, 1).Each(func(i int, el *goquery.Selection) {
				blue, _ = el.Attr("data-teamhighlight")
				if el.HasClass("matchlist-winner-team") {
					winner = teams[blue]
				}
			})

			m := Match{teams[blue], teams[red], winner, false}
			s.matches = append(s.matches, &m)
			if winner != nil {
				winner.wins++
				m.GetLoser().losses++
			}
		})
	}

	return s
}

func SmartFormat(n float64) string {
	if n >= 1.0 {
		return fmt.Sprintf("%.0f%%", n)
	} else if n >= 0.1 {
		return fmt.Sprintf("%.1f%%", n)
	} else if n >= 0.01 {
		return fmt.Sprintf("%.2f%%", n)
	} else {
		return fmt.Sprintf("%.3f%%", n)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
