package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
)

func main() {
	urlconf := map[string]string{
		"lck": "https://lol.gamepedia.com/LCK/2019_Season/Summer_Season",
		"lec": "https://lol.gamepedia.com/LEC/2019_Season/Summer_Season",
		"lcs": "https://lol.gamepedia.com/LCS/2019_Season/Summer_Season",
		"lpl": "https://lol.gamepedia.com/LPL/2019_Season/Summer_Season",
		"lms": "https://lol.gamepedia.com/LMS/2019_Season/Summer_Season",
	}
	teamconf := map[string]map[string]*Team{
		"lck": GetLCKTeams(),
		"lcs": GetLCSTeams(),
		"lec": GetLECTeams(),
		"lpl": GetLPLTeams(),
		"lms": GetLMSTeams(),
	}
	outro := strings.Replace(
		"\n\n Percentages assume each match is a 50/50 tossup\n\n"+
			" The percentages are imperfect due to how tiebreaker math works out; they are there simply for estimates\n\n"+
			" Written in some very low quality Go, pull requests welcome, PM me for link\n\n",
		// "[ Foldy sheet by Adamgo83](https://www.reddit.com/r/leagueoflegends/comments/ciq0sj/clutch_gaming_vs_counter_logic_gaming_lcs_2019/ev8h337/)",
		" ", " ^^^", -1)

	league := os.Args[1]
	markdown := len(os.Args) > 2 && os.Args[2] == "--md"
	if !markdown {
		outro = ""
	}
	displayPct := (len(os.Args) > 2 && os.Args[2] == "--pct") ||
		(len(os.Args) > 3 && (os.Args[2] == "--pct" || os.Args[3] == "--pct"))

	chosen_teams := teamconf[league]
	chosen_url := urlconf[league]
	fmt.Printf("\n\tSimulating %s from %s\n\n", strings.ToUpper(league), chosen_url)
	league_size := len(chosen_teams)

	original_ranking_map := make(map[string]int)
	original_records := make(map[string]string)

	s := ParseSchedule(chosen_url, chosen_teams)

	forces := [][]string{
		// =========================== //
		// ----------- LCS ----------- //
		[]string{"CG", "FLY", "CG"},
		// ----------- LCS ----------- //
		// =========================== //
		// ----------- LEC ----------- //
		// G2
		[]string{"SK", "G2", "G2"},
		[]string{"G2", "VIT", "G2"},
		[]string{"G2", "RGE", "G2"},
		[]string{"G2", "S04", "G2"},
		[]string{"G2", "XL", "G2"},
		[]string{"MSF", "G2", "G2"},
		// FNC
		[]string{"XL", "FNC", "FNC"},
		[]string{"FNC", "VIT", "FNC"},
		[]string{"FNC", "RGE", "FNC"},
		// OG
		[]string{"RGE", "OG", "OG"},
		[]string{"OG", "XL", "OG"},
		// ----------- LEC ----------- //
		// =========================== //
		// ----------- LCK ----------- //
		// // JAG
		// []string{"DWG", "JAG", "DWG"},
		// []string{"JAG", "GEN", "GEN"},
		// []string{"JAG", "KZ", "KZ"},
		// []string{"KT", "JAG", "KT"},
		// []string{"GRF", "JAG", "GRF"},
		// // HLE
		// []string{"KT", "HLE", "KT"},
		// []string{"HLE", "DWG", "DWG"},
		// []string{"HLE", "SKT", "SKT"},
		// []string{"AF", "HLE", "AF"},
		// []string{"HLE", "GRF", "GRF"},
		// ----------- LCK ----------- //
		// =========================== //
	}

	nSim := 0
	var toSkip []int
	for i, m := range s.matches {
		if m.winner != nil {
			continue
		}

		foundForce := false
		for _, force := range forces {
			if force[0] == m.blue.name && force[1] == m.red.name {
				if force[0] == force[2] {
					fmt.Printf("Forcing %v 1-0 %v\n", force[0], force[1])
					m.winner = m.blue
					m.red.losses++
				} else {
					fmt.Printf("Forcing %v 0-1 %v\n", force[0], force[1])
					m.winner = m.red
					m.blue.losses++
				}
				m.winner.wins++
				foundForce = true
				m.simmed = true
			}
		}

		if !foundForce {
			nSim++
		} else {
			toSkip = append(toSkip, i)
		}
	}

	season := Season{}
	for _, t := range chosen_teams {
		season.standings = append(season.standings, Rank{t, false})
	}
	season.schedule = s
	season.Sort()
	// os.Exit(1)
	for i, v := range season.standings {
		// fmt.Printf("%s is %d-%d\n", v.team.name, v.team.wins, v.team.losses)
		original_ranking_map[v.team.name] = i
		original_records[v.team.name] = fmt.Sprintf("%v-%v", v.team.wins, v.team.losses)
	}

	var wg sync.WaitGroup

	finishesUntied := make(map[string]map[int]int)
	finishesTied := make(map[string]map[int]int)

	fmt.Printf("\n\tSimulating %d matches (%v combinations)\n\n",
		nSim, humanize.Comma(int64(math.Pow(2, float64(nSim)))))

	totals := make(map[int]int)
	for combo := range GenerateCombinations("br", nSim) {
		wg.Add(1)
		for newSeason := range ProcessResults(combo, &wg, season, len(season.schedule.matches)-nSim-len(toSkip), forces, toSkip) {
			for i, t := range newSeason.standings {
				if _, ok := totals[i]; !ok {
					totals[i] = 0
				}
				if _, ok := finishesUntied[t.team.name]; !ok {
					finishesUntied[t.team.name] = make(map[int]int)
					finishesTied[t.team.name] = make(map[int]int)
					for n := 0; n < league_size; n++ {
						finishesUntied[t.team.name][n] = 0
						finishesTied[t.team.name][n] = 0
					}
				}

				if t.tie {
					for j := 0; j < league_size; j++ {
						if newSeason.standings[j].tie && newSeason.standings[j].team.wins == t.team.wins {
							finishesTied[t.team.name][j]++
							totals[j]++
						}
					}
				} else {
					finishesUntied[t.team.name][i]++
					totals[i]++
				}
			}
		}
	}

	wg.Wait()

	data := [][]string{}
	teams := []string{}
	for t, _ := range original_ranking_map {
		teams = append(teams, t)
	}
	sort.Slice(teams, func(i, j int) bool { return original_ranking_map[teams[i]] < original_ranking_map[teams[j]] })

	for _, team := range teams {
		counts := finishesUntied[team]
		row := make([]string, league_size+2)
		row[0] = team
		row[1] = original_records[team]

		keys := make([]int, 0, len(counts))
		for key := range counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, key := range keys {
			if counts[key] == 0 && finishesTied[team][key] == 0 {
				// fmt.Printf("%v cannot finish #%v\n", team, key+1)
				row[key+2] = "X"
			} else if displayPct {
				val := (float64(counts[key]) + float64(finishesTied[team][key])) * 100.0 / float64(totals[key])
				row[key+2] = SmartFormat(val)
			}
		}

		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"Team", ""}
	for i := 1; i <= league_size; i++ {
		header = append(header, humanize.Ordinal(i))
	}
	table.SetHeader(header)
	if markdown {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	}

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
	fmt.Println(outro)
}

func GenerateCombinations(alphabet string, length int) <-chan string {
	c := make(chan string)
	go func(c chan string) {
		defer close(c)
		GenerateHelper(c, "", alphabet, length)
	}(c)

	return c
}

func GenerateHelper(c chan string, combo string, alphabet string, length int) {
	if length <= 0 {
		c <- combo
		return
	}

	var newCombo string
	for _, ch := range alphabet {
		newCombo = combo + string(ch)
		GenerateHelper(c, newCombo, alphabet, length-1)
	}
}

func ProcessResults(combination string, wg *sync.WaitGroup, s Season, offset int, forces [][]string, forced []int) <-chan Season {
	c := make(chan Season)
	go func(c chan Season) {
		defer close(c)
		ProcessResultsHelper(c, combination, wg, s, offset, forces, forced)
	}(c)

	return c
}

func ProcessResultsHelper(c chan Season, combination string, wg *sync.WaitGroup,
	s Season, offset int, forces [][]string, forced []int) {
	latest := make(map[string]*Team)
	print := combination == ""
	if !print {
		// return
	}

	for _, r := range s.standings {
		latest[r.team.name] = r.team
	}
	defer wg.Done()

	newSchedule := Schedule{}
	newMatches := make([]*Match, 0)
	for _, m := range s.schedule.matches {
		mCopy := *m
		newMatches = append(newMatches, &mCopy)
		if print {
			fmt.Printf("[%v, %v] [%v, %v]\n", m, &m, mCopy, &mCopy)
		}
	}
	newSchedule.matches = newMatches

	skip := 0
	for x := 0; x < len(combination); x++ {
		winnerColor := string(combination[x])
		if print {
			fmt.Printf("x := %v, skip := %v | offset := %v | len(m) := %v\n", x, skip, offset, len(newSchedule.matches))
		}
		simmedMatch := newSchedule.matches[x+offset+skip]
		if print {
			fmt.Println(simmedMatch)
		}
		if simmedMatch.winner != nil {
			x--
			skip++
			continue
		}

		simmedMatch.simmed = true

		if _, ok := latest[simmedMatch.red.name]; !ok {
			latest[simmedMatch.red.name] = simmedMatch.red
		}
		if _, ok := latest[simmedMatch.blue.name]; !ok {
			latest[simmedMatch.blue.name] = simmedMatch.blue
		}

		simmedMatch.red = latest[simmedMatch.red.name]
		simmedMatch.blue = latest[simmedMatch.blue.name]

		blueCopy := *simmedMatch.blue
		redCopy := *simmedMatch.red

		simmedMatch.blue = &blueCopy
		simmedMatch.red = &redCopy
		latest[simmedMatch.blue.name] = &blueCopy
		latest[simmedMatch.red.name] = &redCopy

		for _, forceMatch := range forces {
			if simmedMatch.blue.name == forceMatch[0] && simmedMatch.red.name == forceMatch[1] {
				if forceMatch[0] == forceMatch[2] {
					winnerColor = "b"
				} else {
					winnerColor = "r"
				}
				simmedMatch.simmed = true
			}
		}

		if winnerColor == "b" {
			simmedMatch.winner = simmedMatch.blue
		} else {
			simmedMatch.winner = simmedMatch.red
		}

		simmedMatch.winner.wins += 1
		simmedMatch.GetLoser().losses += 1
	}

	newSeason := Season{}
	for _, v := range latest {
		newSeason.standings = append(newSeason.standings, Rank{v, false})
	}
	newSeason.schedule = newSchedule

	checkForTeam := ""
	checkForFinish := 4
	checkQuietly := true

	newSeason.Sort()

	if checkForTeam != "" {
		if newSeason.standings[checkForFinish-1].team.name == checkForTeam {
			if checkQuietly {
				fmt.Printf("! Found %v in %v\n", checkForTeam, humanize.Ordinal(checkForFinish))
			} else {
				fmt.Printf("%v finishes %v when the next %v matches are won by %v. Final standings:\n%v\n",
					checkForTeam, humanize.Ordinal(checkForFinish), len(combination), combination, newSeason)

				skip = 0
				for x := 0; x < len(combination); x++ {
					winnerColor := string(combination[x])
					simmedMatch := s.schedule.matches[x+offset+skip]
					if simmedMatch.winner != nil {
						x--
						skip++
						continue
					}

					if winnerColor == "r" {
						fmt.Printf("Match #%v: %v 0-1 %v\n", x+offset+skip, simmedMatch.blue.name, simmedMatch.red.name)
					} else {
						fmt.Printf("Match #%v: %v 1-0 %v\n", x+offset+skip, simmedMatch.blue.name, simmedMatch.red.name)
					}
				}
				fmt.Println("===============")
			}
		}
	}

	c <- newSeason
}
