package game

type Column struct {
	ID   string // e.g. "t2b", "b2t", "free", "announce"
	Name string // user-facing
}

type Row struct {
	ID   string // e.g. "ones", "twos", "yamb"
	Name string
}

type ScoreCard struct {
	Rows    []Row
	Columns []Column
	Scores  map[string]map[string]*int // rowID -> colID -> score
}

func NewScoreCard() ScoreCard {
	cols := []Column{
		{ID: "t2b", Name: "⬇️"},
		{ID: "b2t", Name: "⬆️"},
		{ID: "free", Name: "🔃"},
		{ID: "announce", Name: "N"},
	}

	rows := []Row{
		{ID: "1", Name: "1"},
		{ID: "2", Name: "2"},
		{ID: "3", Name: "3"},
		{ID: "4", Name: "4"},
		{ID: "5", Name: "5"},
		{ID: "6", Name: "6"},
		{ID: "sum1", Name: "Sum"},
		{ID: "max", Name: "Max"},
		{ID: "min", Name: "Min"},
		{ID: "sum2", Name: "Sum"},
		{ID: "kenta", Name: "Kenta"},
		{ID: "full", Name: "Full"},
		{ID: "yamb", Name: "Yamb"},
		{ID: "sum3", Name: "Sum"},
	}

	// initialize empty scores
	scores := map[string]map[string]*int{}
	for _, r := range rows {
		scores[r.ID] = map[string]*int{}
		for _, c := range cols {
			scores[r.ID][c.ID] = nil
		}
	}

	return ScoreCard{
		Rows:    rows,
		Columns: cols,
		Scores:  scores,
	}
}
