package questions

type Question struct {
	ID          int     `json:"id"`
	Text        string  `json:"text"`
	TargetValue float64 `json:"target_value"`
	Unit        string  `json:"unit"`
}

// Bank holds 50 questions split between NHL and Champions League events.
// TargetValue is the authoritative answer revealed only when the host calls the round.
var Bank = []Question{
	// NHL
	{1, "How many points will McDavid record in tonight's game?", 3, "points"},
	{2, "How many goals will the Oilers score tonight?", 4, "goals"},
	{3, "How many saves will the opposing goalie make tonight?", 28, "saves"},
	{4, "How many penalty minutes will be assessed tonight?", 10, "minutes"},
	{5, "How many shots on goal will McDavid take tonight?", 5, "shots"},
	{6, "How many power play goals will the Oilers score tonight?", 2, "goals"},
	{7, "What will the total number of goals be in tonight's game?", 6, "goals"},
	{8, "How many hits will the Oilers register tonight?", 22, "hits"},
	{9, "How many faceoffs will McDavid win tonight?", 14, "faceoffs"},
	{10, "How many blocked shots will the Oilers have tonight?", 12, "blocks"},
	{11, "How many points will Draisaitl record this week?", 5, "points"},
	{12, "How many goals will Matthews score this month?", 8, "goals"},
	{13, "What will the winning team's margin of victory be tonight?", 2, "goals"},
	{14, "How many assists will Makar record tonight?", 2, "assists"},
	{15, "How many shots on goal will be taken in the first period tonight?", 18, "shots"},
	{16, "How many minutes of ice time will McDavid log tonight?", 22, "minutes"},
	{17, "How many giveaways will the home team commit tonight?", 9, "giveaways"},
	{18, "How many takeaways will the away team record tonight?", 7, "takeaways"},
	{19, "How many goals will be scored in overtime or shootout tonight?", 1, "goals"},
	{20, "How many penalty shots will be awarded in tonight's game?", 0, "shots"},
	{21, "How many goals will Ovechkin score this week?", 3, "goals"},
	{22, "What will Fleury's save percentage be tonight (out of 100)?", 92, "percent"},
	{23, "How many hat tricks will be scored in tonight's game?", 0, "hat tricks"},
	{24, "How many wins will the Avalanche have this month?", 12, "wins"},
	{25, "How many points will the Bruins top scorer get this week?", 4, "points"},
	// Champions League
	{26, "How many goals will Mbappe score in tonight's CL match?", 2, "goals"},
	{27, "How many corners will be taken in tonight's CL match?", 9, "corners"},
	{28, "How many yellow cards will be shown in tonight's match?", 3, "cards"},
	{29, "How many shots on target will the home team have tonight?", 7, "shots"},
	{30, "How many saves will the goalkeeper make tonight?", 5, "saves"},
	{31, "How many fouls will be committed in tonight's CL match?", 24, "fouls"},
	{32, "How many offsides will be called in tonight's match?", 4, "offsides"},
	{33, "How many passes will Bellingham complete tonight?", 58, "passes"},
	{34, "How many minutes until the first goal in tonight's match?", 23, "minutes"},
	{35, "How many goals will be scored in tonight's CL match total?", 3, "goals"},
	{36, "How many substitutions will Real Madrid make tonight?", 3, "substitutions"},
	{37, "How many km will Vinicius Jr run in tonight's match?", 11, "km"},
	{38, "How many successful dribbles will Vinicius Jr complete tonight?", 6, "dribbles"},
	{39, "How many crosses will be attempted in tonight's CL match?", 18, "crosses"},
	{40, "What will the xG total for both teams be tonight (×10)?", 25, "xG×10"},
	{41, "How many red cards will be shown in tonight's CL match?", 0, "cards"},
	{42, "How many goals will Haaland score in tonight's CL match?", 2, "goals"},
	{43, "How many tackles won will the away team register tonight?", 14, "tackles"},
	{44, "How many minutes will be added in stoppage time in the second half?", 5, "minutes"},
	{45, "How many VAR reviews will occur in tonight's match?", 2, "reviews"},
	{46, "How many goals will be scored from set pieces tonight?", 1, "goals"},
	{47, "How many aerial duels will be won by the home team tonight?", 12, "duels"},
	{48, "How many touches in the box will the top scorer have tonight?", 8, "touches"},
	{49, "How many goal-line clearances will there be in tonight's match?", 1, "clearances"},
	{50, "How many matches in tonight's CL fixtures will end in a draw?", 2, "matches"},
}
