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
	}
	teamconf := map[string]map[string]*Team{
		"lck": GetLCKTeams(),
		"lcs": GetLCSTeams(),
		"lec": GetLECTeams(),
		"lpl": GetLPLTeams(),
	}

	league := os.Args[1]
	chosen_teams := teamconf[league]
	chosen_url := urlconf[league]
	fmt.Printf("\n\tSimulating %s from %s\n\n", strings.ToUpper(league), chosen_url)
	league_size := len(chosen_teams)

	original_ranking_map := make(map[string]int)
	original_records := make(map[string]string)

	s := ParseSchedule(chosen_url, chosen_teams)

	forces := [][]string{
		// ----------- LCS ----------- //
		// []string{"FLY", "C9", "C9"},
		// []string{"C9", "TL", "TL"},
		// []string{"TSM", "FOX", "FOX"},
		// []string{"TL", "OPT", "OPT"},
		// []string{"FLY", "C9", "FLY"},
		// ----------- LCS ----------- //

		// ----------- LEC ----------- //
		// G2
		[]string{"G2", "VIT", "G2"},
		[]string{"G2", "RGE", "G2"},
		[]string{"G2", "S04", "G2"},
		[]string{"G2", "XL", "G2"},
		[]string{"G2", "OG", "G2"},
		// ----------- LEC ----------- //

		// ----------- LCK ----------- //
		// JAG
		[]string{"DWG", "JAG", "DWG"},
		[]string{"JAG", "GEN", "GEN"},
		[]string{"JAG", "KZ", "KZ"},
		[]string{"KT", "JAG", "KT"},
		[]string{"GRF", "JAG", "GRF"},
		// HLE
		[]string{"KT", "HLE", "KT"},
		[]string{"HLE", "DWG", "DWG"},
		[]string{"HLE", "SKT", "SKT"},
		[]string{"AF", "HLE", "AF"},
		[]string{"HLE", "GRF", "GRF"},
		// ----------- LCK ----------- //
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
					m.winner = m.blue
				} else {
					m.winner = m.red
				}
				foundForce = true
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
	season.Sort()
	for i, v := range season.standings {
		// fmt.Printf("%s is %d-%d\n", v.team.name, v.team.wins, v.team.losses)
		original_ranking_map[v.team.name] = i
		original_records[v.team.name] = fmt.Sprintf("%v-%v", v.team.wins, v.team.losses)
	}
	season.schedule = s

	var wg sync.WaitGroup

	finishes := make(map[string]map[int]int)

	fmt.Printf("\n\tSimulating %d matches (%v combinations)\n\n",
		nSim, humanize.Comma(int64(math.Pow(2, float64(nSim)))))
	for combo := range GenerateCombinations("br", nSim) {
		wg.Add(1)
		for newSeason := range ProcessResults(combo, &wg, season, len(season.schedule.matches)-nSim, forces, toSkip) {
			for i, t := range newSeason.standings {
				if _, ok := finishes[t.team.name]; !ok {
					finishes[t.team.name] = make(map[int]int)
					for n := 0; n < league_size; n++ {
						finishes[t.team.name][n] = 0
					}
				}

				finishes[t.team.name][i] += 1

				for j := 0; j < league_size; j++ {
					if newSeason.standings[j].tie && newSeason.standings[j].team.wins == t.team.wins {
						finishes[t.team.name][j] += 1
					}
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
		counts := finishes[team]
		// row := []string{team, original_records[team], "", "", "", "", "", "", "", "", "", ""}

		row := make([]string, league_size+2)
		row[0] = team
		row[1] = original_records[team]

		keys := make([]int, 0, len(counts))
		for key := range counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, key := range keys {
			if counts[key] == 0 {
				// fmt.Printf("%v cannot finish #%v\n", team, key+1)
				row[key+2] = "X"
			}
		}

		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	// table.SetHeader([]string{"Team", "", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th"})

	header := []string{"Team", ""}
	for i := 1; i <= league_size; i++ {
		header = append(header, humanize.Ordinal(i))
	}
	table.SetHeader(header)

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
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

	for _, r := range s.standings {
		latest[r.team.name] = r.team
	}

	defer wg.Done()
	for i, _ := range combination {
		winnerColor := string(combination[i])
		simmedMatch := s.schedule.matches[i+offset]
		if simmedMatch.winner != nil {
			continue
		}

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

	newSeason.Sort()
	c <- newSeason
}
