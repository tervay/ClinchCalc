package main

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/olekukonko/tablewriter"
)

func main() {
	lcs := GetLCSTeams()

	original_ranking_map := make(map[string]int)
	original_records := make(map[string]string)

	s := ParseSchedule("https://lol.gamepedia.com/LCS/2019_Season/Summer_Season", lcs)

	forces := [][]string{
		// []string{"TSM", "FOX", "TSM"},
		// []string{"FOX", "CLG", "CLG"},
		// []string{"TL", "FOX", "TL"},
		// []string{"FOX", "FLY", "FLY"},
	}

	nSim := 0
	for _, m := range s.matches {
		if nSim > 0 && m.winner != nil {
			fmt.Println("skip")
		}

		if m.winner != nil {
			continue
		}
		nSim++
	}

	season := Season{}
	for _, t := range lcs {
		season.standings = append(season.standings, Rank{t, false})
	}
	season.Sort()
	for i, v := range season.standings {
		fmt.Printf("%s is %d-%d\n", v.team.name, v.team.wins, v.team.losses)
		original_ranking_map[v.team.name] = i
		original_records[v.team.name] = fmt.Sprintf("%v-%v", v.team.wins, v.team.losses)
	}
	season.schedule = s

	var wg sync.WaitGroup

	finishes := make(map[string]map[int]int)

	for combo := range GenerateCombinations("br", nSim) {
		wg.Add(1)
		for newSeason := range ProcessResults(combo, &wg, season, len(season.schedule.matches)-nSim, forces) {
			for i, t := range newSeason.standings {
				if _, ok := finishes[t.team.name]; !ok {
					finishes[t.team.name] = make(map[int]int)
					for n := range [10]int{} {
						finishes[t.team.name][n] = 0
					}
				}

				finishes[t.team.name][i] += 1

				j := 0
				for j < 10 {
					if newSeason.standings[j].tie && newSeason.standings[j].team.wins == t.team.wins {
						finishes[t.team.name][j] += 1
					}
					j++
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
		row := []string{team, original_records[team], "", "", "", "", "", "", "", "", "", ""}
		keys := make([]int, 0, len(counts))
		for key := range counts {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		for _, key := range keys {
			if counts[key] == 0 {
				fmt.Printf("%v cannot finish #%v\n", team, key+1)
				row[key+2] = "X"
			}
		}

		data = append(data, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Team", "", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th"})
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

func ProcessResults(combination string, wg *sync.WaitGroup, s Season, offset int, forces [][]string) <-chan Season {
	c := make(chan Season)
	go func(c chan Season) {
		defer close(c)
		ProcessResultsHelper(c, combination, wg, s, offset, forces)
	}(c)

	return c
}

func ProcessResultsHelper(c chan Season, combination string, wg *sync.WaitGroup, s Season, offset int, forces [][]string) {
	latest := make(map[string]*Team)
	defer wg.Done()
	for i, _ := range combination {
		winnerColor := string(combination[i])
		simmedMatch := s.schedule.matches[i+offset]
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
