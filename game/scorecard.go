package game

import "errors"

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
	// *int to allow nil (unfilled) scores
	Scores map[string]map[string]*int // rowID -> colID -> score
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
		{ID: "poker", Name: "Poker"},
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

func (sc *ScoreCard) fillT2B(rowID string, score int) error {
	// check if the field above is filled (if not the first row)
	for i, r := range sc.Rows {
		if r.ID == rowID {
			if i == 0 {
				// first row, no need to check above
				sc.Scores[rowID]["t2b"] = &score
				return nil
			}
			aboveRowID := sc.Rows[i-1].ID
			// if the above row is a sum, check the row above that (since sums are not filled by players)
			if len(aboveRowID) >= 3 && aboveRowID[:3] == "sum" {
				aboveRowID = sc.Rows[i-2].ID
			}
			if sc.Scores[aboveRowID]["t2b"] == nil {
				return errors.New("field above is not filled")
			}
			sc.Scores[rowID]["t2b"] = &score
			return nil
		}
	}
	return errors.New("unknown row ID")
}

func (sc *ScoreCard) fillB2T(rowID string, score int) error {
	// check if the field below is filled (if not the last row)
	for i, r := range sc.Rows {
		if r.ID == rowID {
			if i == len(sc.Rows)-1 {
				// last row, no need to check below
				sc.Scores[rowID]["b2t"] = &score
				return nil
			}
			belowRowID := sc.Rows[i+1].ID
			// if the below row is a sum, check the row below that (since sums are not filled by players)
			if len(belowRowID) >= 3 && belowRowID[:3] == "sum" {
				belowRowID = sc.Rows[i+2].ID
			}
			if sc.Scores[belowRowID]["b2t"] == nil {
				return errors.New("field below is not filled")
			}
			sc.Scores[rowID]["b2t"] = &score
			return nil
		}
	}
	return errors.New("unknown row ID")
}

func (sc *ScoreCard) fillFree(rowID string, score int) error {
	sc.Scores[rowID]["free"] = &score
	return nil
}

func (sc *ScoreCard) FillField(rowID, colID string, dice *Dice) (int, error) {
	if sc.Scores[rowID][colID] != nil {
		return 0, errors.New("field already filled")
	}

	score, err := sc.CalculateScore(rowID, dice)
	if err != nil {
		return 0, err
	}

	switch colID {
	case "t2b":
		return score, sc.fillT2B(rowID, score)
	case "b2t":
		return score, sc.fillB2T(rowID, score)
	case "free":
		return score, sc.fillFree(rowID, score)
	case "announce":
		return score, errors.New("TODO")
	}

	return 0, errors.New("unknown column ID")
}

// TODO: make error messages more informative
func (sc *ScoreCard) CalculateScore(rowID string, dice *Dice) (int, error) {
	switch rowID {
	case "1":
		return dice.Count(1) * 1, nil
	case "2":
		return dice.Count(2) * 2, nil
	case "3":
		return dice.Count(3) * 3, nil
	case "4":
		return dice.Count(4) * 4, nil
	case "5":
		return dice.Count(5) * 5, nil
	case "6":
		return dice.Count(6) * 6, nil
	case "max":
		val := dice.MinMax()
		if val == 0 {
			return 0, errors.New("no max")
		}
		return val, nil
	case "min":
		val := dice.MinMax()
		if val == 0 {
			return 0, errors.New("no min")
		}
	case "kenta":
		val := dice.Kenta()
		if val == 0 {
			return 0, errors.New("no kenta")
		}
		return val, nil
	case "full":
		val := dice.Full()
		if val == 0 {
			return 0, errors.New("no full")
		}
		return val, nil
	case "poker":
		val := dice.Poker()
		if val == 0 {
			return 0, errors.New("no poker")
		}
		return val, nil
	case "yamb":
		val := dice.Yamb()
		if val == 0 {
			return 0, errors.New("no yamb")
		}
		return val, nil
	}
	return 0, errors.New("unknown row ID")
}
