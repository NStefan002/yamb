package game

import "errors"

type Column struct {
	ID   string // e.g. "t2b", "b2t", "free", "announce"
	Name string // user-facing
}

type Row struct {
	ID   string // e.g. "1", "2", "yamb"
	Name string
}

type ScoreCard struct {
	Rows    []Row
	Columns []Column
	// *int to allow nil (unfilled) scores
	Scores map[string]map[string]*int // rowID -> colID -> score
	// one singular selected cell for highlighting in UI (rowID, colID)
	SelectedCell [2]string
}

func NewScoreCard() ScoreCard {
	cols := []Column{
		{ID: "t2b", Name: "↓"},
		{ID: "b2t", Name: "↑"},
		{ID: "free", Name: "↑↓"},
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
		Rows:         rows,
		Columns:      cols,
		Scores:       scores,
		SelectedCell: [2]string{"", ""}, // no cell selected initially
	}
}

func (sc *ScoreCard) SelectCell(rowID, colID string) error {
	if len(rowID) >= 3 && rowID[:3] == "sum" {
		// cannot select sum rows
		return errors.New("cannot select sum rows")
	}

	if sc.Scores[rowID][colID] != nil {
		// if cell is already filled, cannot select
		return errors.New("cannot select filled cell")
	}

	if sc.SelectedCell[0] == rowID && sc.SelectedCell[1] == colID {
		// unselect if already selected
		sc.SelectedCell[0] = ""
		sc.SelectedCell[1] = ""
	} else {
		sc.SelectedCell[0] = rowID
		sc.SelectedCell[1] = colID
	}

	return nil
}

func (sc *ScoreCard) GetSelectedCell() (string, string) {
	return sc.SelectedCell[0], sc.SelectedCell[1]
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
			if i == len(sc.Rows)-2 { // -2 because actual last row is sum
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
		return dice.Number(1)
	case "2":
		return dice.Number(2)
	case "3":
		return dice.Number(3)
	case "4":
		return dice.Number(4)
	case "5":
		return dice.Number(5)
	case "6":
		return dice.Number(6)
	case "max":
		return dice.MinMax()
	case "min":
		return dice.MinMax()
	case "kenta":
		return dice.Kenta()
	case "full":
		return dice.Full()
	case "poker":
		return dice.Poker()
	case "yamb":
		return dice.Yamb()
	}
	return 0, errors.New("unknown row ID")
}

// sum of 1-6 plus 30 if >= 60
func (sc *ScoreCard) calcSum1(colID string) int {
	sum := 0
	allFilled := true
	for _, r := range sc.Rows {
		if r.ID == "sum1" {
			break
		}
		if sc.Scores[r.ID][colID] != nil {
			sum += *sc.Scores[r.ID][colID]
		} else {
			allFilled = false
		}
	}
	if allFilled {
		if sum >= 60 {
			sum += 30
		}
		return sum
	}
	return 0
}

// (max - min) * 1s
func (sc *ScoreCard) calcSum2(colID string) int {
	if sc.Scores["max"][colID] == nil || sc.Scores["min"][colID] == nil || sc.Scores["1"][colID] == nil {
		return 0
	}
	return (*sc.Scores["max"][colID] - *sc.Scores["min"][colID]) * *sc.Scores["1"][colID]
}

func (sc *ScoreCard) calcSum3(colID string) int {
	sum := 0
	allFilled := true
	started := false
	for _, r := range sc.Rows {
		if r.ID == "sum3" {
			break
		}
		if r.ID == "sum2" {
			started = true
			continue
		}
		if started {
			if sc.Scores[r.ID][colID] != nil {
				sum += *sc.Scores[r.ID][colID]
			} else {
				allFilled = false
			}
		}
	}
	if allFilled {
		return sum
	}
	return 0
}

func (sc *ScoreCard) CalculateSums() {
	for _, col := range sc.Columns {
		if sc.Scores["sum1"][col.ID] == nil {
			sum := sc.calcSum1(col.ID)
			if sum > 0 {
				sc.Scores["sum1"][col.ID] = &sum
			}
		}
		if sc.Scores["sum2"][col.ID] == nil {
			sum := sc.calcSum2(col.ID)
			if sum > 0 {
				sc.Scores["sum2"][col.ID] = &sum
			}
		}
		if sc.Scores["sum3"][col.ID] == nil {
			sum := sc.calcSum3(col.ID)
			if sum > 0 {
				sc.Scores["sum3"][col.ID] = &sum
			}
		}
	}
}
