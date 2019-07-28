### Installation

```bash
λ go get github.com/dustin/go-humanize
λ go get github.com/olekukonko/tablewriter
```

# Example usage

```bash
λ go run main.go data.go <league> [--md] [--pct]
```

where `league` is one of lcs, lec, lpl, lck, or lms (lowercase, case-sensitive). `--md` changes the table to be a markdown table and has a small footer message attached. `--pct` displays the percentages in the table.

If you want the simulator to assume a certain match is going to be won by a specific team, you can specify it in `main.go` via the `forces` map -- keep in mind it's blue/red side sensitive, so check gamepedia beforehand.

```bash
λ go run main.go data.go lcs

        Simulating LCS from https://lol.gamepedia.com/LCS/2019_Season/Summer_Season


        Simulating 15 matches (32,768 combinations)

+------+------+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+
| TEAM |      | 1ST | 2ND | 3RD | 4TH | 5TH | 6TH | 7TH | 8TH | 9TH | 10TH |
+------+------+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+
| TL   | 12-3 |     |     |     |     | X   | X   | X   | X   | X   | X    |
| CLG  | 10-5 |     |     |     |     |     |     | X   | X   | X   | X    |
| TSM  | 9-6  |     |     |     |     |     |     |     | X   | X   | X    |
| C9   | 9-6  |     |     |     |     |     |     |     | X   | X   | X    |
| OPT  | 8-7  | X   |     |     |     |     |     |     |     |     | X    |
| GGS  | 7-8  | X   |     |     |     |     |     |     |     |     | X    |
| 100T | 6-9  | X   | X   |     |     |     |     |     |     |     |      |
| CG   | 6-9  | X   | X   |     |     |     |     |     |     |     |      |
| FLY  | 5-10 | X   | X   | X   | X   |     |     |     |     |     |      |
| FOX  | 3-12 | X   | X   | X   | X   | X   | X   | X   |     |     |      |
+------+------+-----+-----+-----+-----+-----+-----+-----+-----+-----+------+

λ go run main.go data.go lcs --md

        Simulating LCS from https://lol.gamepedia.com/LCS/2019_Season/Summer_Season


        Simulating 15 matches (32,768 combinations)

| TEAM |      | 1ST | 2ND | 3RD | 4TH | 5TH | 6TH | 7TH | 8TH | 9TH | 10TH |
|------|------|-----|-----|-----|-----|-----|-----|-----|-----|-----|------|
| TL   | 12-3 |     |     |     |     | X   | X   | X   | X   | X   | X    |
| CLG  | 10-5 |     |     |     |     |     |     | X   | X   | X   | X    |
| TSM  | 9-6  |     |     |     |     |     |     |     | X   | X   | X    |
| C9   | 9-6  |     |     |     |     |     |     |     | X   | X   | X    |
| OPT  | 8-7  | X   |     |     |     |     |     |     |     |     | X    |
| GGS  | 7-8  | X   |     |     |     |     |     |     |     |     | X    |
| 100T | 6-9  | X   | X   |     |     |     |     |     |     |     |      |
| CG   | 6-9  | X   | X   |     |     |     |     |     |     |     |      |
| FLY  | 5-10 | X   | X   | X   | X   |     |     |     |     |     |      |
| FOX  | 3-12 | X   | X   | X   | X   | X   | X   | X   |     |     |      |


 ^^^Curious ^^^about ^^^a ^^^universe ^^^where ^^^your ^^^favorite ^^^team ^^^finishes ^^^in ^^^X ^^^position? ^^^Let ^^^me ^^^know!

 ^^^This ^^^does ^^^not ^^^account ^^^for ^^^head-to-head ^^^tiebreakers ^^^:( ^^^Code ^^^is ^^^hard

 ^^^Written ^^^in ^^^some ^^^very ^^^low ^^^quality ^^^Go, ^^^pull ^^^requests ^^^welcome, ^^^PM ^^^me ^^^for ^^^link

λ go run main.go data.go lcs --md --pct

        Simulating LCS from https://lol.gamepedia.com/LCS/2019_Season/Summer_Season


        Simulating 15 matches (32,768 combinations)

| TEAM |      | 1ST | 2ND  |  3RD  | 4TH  | 5TH  | 6TH  | 7TH  | 8TH |  9TH  | 10TH |
|------|------|-----|------|-------|------|------|------|------|-----|-------|------|
| TL   | 12-3 | 86% | 8%   | 2%    | 0.2% | X    | X    | X    | X   | X     | X    |
| CLG  | 10-5 | 10% | 49%  | 23%   | 12%  | 4%   | 0.2% | X    | X   | X     | X    |
| TSM  | 9-6  | 2%  | 18%  | 30%   | 31%  | 16%  | 4%   | 0.8% | X   | X     | X    |
| C9   | 9-6  | 2%  | 18%  | 29%   | 27%  | 18%  | 6%   | 1.0% | X   | X     | X    |
| OPT  | 8-7  | X   | 6%   | 13%   | 21%  | 33%  | 20%  | 8%   | 2%  | 0.06% | X    |
| GGS  | 7-8  | X   | 0.4% | 3%    | 8%   | 20%  | 32%  | 24%  | 12% | 4%    | X    |
| CG   | 6-9  | X   | X    | 0.02% | 0.6% | 5%   | 18%  | 26%  | 28% | 21%   | 1%   |
| 100T | 6-9  | X   | X    | 0.02% | 0.8% | 4%   | 17%  | 29%  | 31% | 18%   | 1%   |
| FLY  | 5-10 | X   | X    | X     | X    | 0.2% | 4%   | 12%  | 26% | 47%   | 14%  |
| FOX  | 3-12 | X   | X    | X     | X    | X    | X    | X    | 1%  | 11%   | 84%  |


 ^^^Percentages ^^^assume ^^^that ^^^each ^^^match ^^^is ^^^a ^^^50/50 ^^^tossup

 ^^^Curious ^^^about ^^^a ^^^universe ^^^where ^^^your ^^^favorite ^^^team ^^^finishes ^^^in ^^^X ^^^position? ^^^Let ^^^me ^^^know!

 ^^^This ^^^does ^^^not ^^^account ^^^for ^^^head-to-head ^^^tiebreakers ^^^:( ^^^Code ^^^is ^^^hard

 ^^^Written ^^^in ^^^some ^^^very ^^^low ^^^quality ^^^Go, ^^^pull ^^^requests ^^^welcome, ^^^PM ^^^me ^^^for ^^^link

λ go run main.go data.go lms

        Simulating LMS from https://lol.gamepedia.com/LMS/2019_Season/Summer_Season


        Simulating 8 matches (256 combinations)

+------+-----+-----+-----+-----+-----+-----+-----+-----+
| TEAM |     | 1ST | 2ND | 3RD | 4TH | 5TH | 6TH | 7TH |
+------+-----+-----+-----+-----+-----+-----+-----+-----+
| JT   | 9-0 |     | X   | X   | X   | X   | X   | X   |
| AHQ  | 7-4 | X   |     |     |     | X   | X   | X   |
| MAD  | 5-4 | X   |     |     |     |     |     | X   |
| HKA  | 5-5 | X   |     |     |     |     |     | X   |
| GRX  | 4-6 | X   | X   |     |     |     |     |     |
| ALF  | 2-7 | X   | X   | X   |     |     |     |     |
| FW   | 2-8 | X   | X   | X   | X   |     |     |     |
+------+-----+-----+-----+-----+-----+-----+-----+-----+
```