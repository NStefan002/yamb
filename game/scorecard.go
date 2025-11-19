package game

import (
	"errors"
)

// column ids

const (
	TopToBottom            string = "t2b"
	BottomToTop            string = "b2t"
	Free                   string = "free"
	Announced              string = "announced"
	MiddleToTopAndToBottom string = "m2tb"
	TopAndBottomToMiddle   string = "tb2m"
	Forced                 string = "forced"
	Hand                   string = "hand"
	Maximum                string = "maximum"
)

// row ids

const (
	Ones      string = "1"
	Twos      string = "2"
	Threes    string = "3"
	Fours     string = "4"
	Fives     string = "5"
	Sixes     string = "6"
	Sum1      string = "sum1"
	Max       string = "max"
	Min       string = "min"
	Sum2      string = "sum2"
	Straight  string = "straight"
	Trips     string = "trips"
	FullHouse string = "fullhouse"
	Quads     string = "quads"
	Yamb      string = "yamb"
	Sum3      string = "sum3"
)

type Column struct {
	ID   string
	Name string // user-facing
}

type Row struct {
	ID   string
	Name string
}

type ScoreCard struct {
	Rows    []Row
	Columns []Column
	// *int to allow nil (unfilled) scores
	Scores map[string]map[string]*int // rowID -> colID -> score
	// one singular selected cell for highlighting in UI (rowID, colID)
	SelectedCell [2]string
	Announced    bool // whether the player has announced their move
}

func NewScoreCard() ScoreCard {
	cols := []Column{
		{ID: TopToBottom, Name: "↓"},
		{ID: BottomToTop, Name: "↑"},
		{ID: Free, Name: "↑↓"},
		{ID: Announced, Name: "A"},
	}

	rows := []Row{
		{ID: Ones, Name: "1"},
		{ID: Twos, Name: "2"},
		{ID: Threes, Name: "3"},
		{ID: Fours, Name: "4"},
		{ID: Fives, Name: "5"},
		{ID: Sixes, Name: "6"},
		{ID: Sum1, Name: "Sum"},
		{ID: Max, Name: "Max"},
		{ID: Min, Name: "Min"},
		{ID: Sum2, Name: "Sum"},
		{ID: Straight, Name: "Straight"},
		{ID: FullHouse, Name: "Full House"},
		{ID: Quads, Name: "Quads"},
		{ID: Yamb, Name: "Yamb"},
		{ID: Sum3, Name: "Sum"},
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
		sc.UnselectCell()
	} else {
		sc.SelectedCell[0] = rowID
		sc.SelectedCell[1] = colID
	}

	return nil
}

func (sc *ScoreCard) UnselectCell() {
	sc.SelectedCell[0] = ""
	sc.SelectedCell[1] = ""
}

func (sc *ScoreCard) GetSelectedCell() (string, string) {
	return sc.SelectedCell[0], sc.SelectedCell[1]
}

func (sc *ScoreCard) Announce() {
	sc.Announced = true
}

func (sc *ScoreCard) IsAnnounced() bool {
	return sc.Announced
}

func (sc *ScoreCard) fillTopToBottom(rowID string, score int) error {
	// check if the field above is filled (if not the first row)
	for i, r := range sc.Rows {
		if r.ID == rowID {
			if i == 0 {
				// first row, no need to check above
				sc.Scores[rowID][TopToBottom] = &score
				return nil
			}
			aboveRowID := sc.Rows[i-1].ID
			// if the above row is a sum, check the row above that (since sums are not filled by players)
			if len(aboveRowID) >= 3 && aboveRowID[:3] == "sum" {
				aboveRowID = sc.Rows[i-2].ID
			}
			if sc.Scores[aboveRowID][TopToBottom] == nil {
				return errors.New("field above is not filled")
			}
			sc.Scores[rowID][TopToBottom] = &score
			return nil
		}
	}
	return errors.New("unknown row ID")
}

func (sc *ScoreCard) fillBottomToTop(rowID string, score int) error {
	// check if the field below is filled (if not the last row)
	for i, r := range sc.Rows {
		if r.ID == rowID {
			if i == len(sc.Rows)-2 { // -2 because actual last row is sum
				// last row, no need to check below
				sc.Scores[rowID][BottomToTop] = &score
				return nil
			}
			belowRowID := sc.Rows[i+1].ID
			// if the below row is a sum, check the row below that (since sums are not filled by players)
			if len(belowRowID) >= 3 && belowRowID[:3] == "sum" {
				belowRowID = sc.Rows[i+2].ID
			}
			if sc.Scores[belowRowID][BottomToTop] == nil {
				return errors.New("field below is not filled")
			}
			sc.Scores[rowID][BottomToTop] = &score
			return nil
		}
	}
	return errors.New("unknown row ID")
}

func (sc *ScoreCard) fillFree(rowID string, score int) error {
	sc.Scores[rowID][Free] = &score
	return nil
}

func (sc *ScoreCard) fillAnnounce(rowID string, score int) error {
	if !sc.Announced {
		return errors.New("must announce before filling this cell")
	}
	sc.Scores[rowID][Announced] = &score
	sc.Announced = false // reset announce after filling
	return nil
}

func (sc *ScoreCard) FillCell(rowID, colID string, dice *Dice) (int, error) {
	if sc.Scores[rowID][colID] != nil {
		return 0, errors.New("field already filled")
	}

	score, err := sc.CalculateScore(rowID, dice)
	if err != nil {
		return 0, err
	}

	switch colID {
	case TopToBottom:
		return score, sc.fillTopToBottom(rowID, score)
	case BottomToTop:
		return score, sc.fillBottomToTop(rowID, score)
	case Free:
		return score, sc.fillFree(rowID, score)
	case Announced:
		return score, sc.fillAnnounce(rowID, score)
	}

	return 0, errors.New("unknown column ID")
}

func (sc *ScoreCard) CalculateScore(rowID string, dice *Dice) (int, error) {
	switch rowID {
	case Ones:
		return dice.Number(1)
	case Twos:
		return dice.Number(2)
	case Threes:
		return dice.Number(3)
	case Fours:
		return dice.Number(4)
	case Fives:
		return dice.Number(5)
	case Sixes:
		return dice.Number(6)
	case Max:
		return dice.MinMax()
	case Min:
		return dice.MinMax()
	case Straight:
		return dice.Kenta()
	case FullHouse:
		return dice.Full()
	case Quads:
		return dice.Poker()
	case Yamb:
		return dice.Yamb()
	}
	return 0, errors.New("unknown row ID")
}

// sum of 1-6 plus 30 if >= 60
func (sc *ScoreCard) calcSum1(colID string) *int {
	sum := 0
	allFilled := true
	for _, r := range sc.Rows {
		if r.ID == Sum1 {
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
		return &sum
	}
	return nil
}

// (max - min) * 1s
func (sc *ScoreCard) calcSum2(colID string) *int {
	if sc.Scores[Max][colID] == nil || sc.Scores[Min][colID] == nil || sc.Scores[Ones][colID] == nil {
		return nil
	}
	sum := (*sc.Scores[Max][colID] - *sc.Scores[Min][colID]) * *sc.Scores[Ones][colID]
	return &sum
}

func (sc *ScoreCard) calcSum3(colID string) *int {
	sum := 0
	allFilled := true
	started := false
	for _, r := range sc.Rows {
		if r.ID == Sum3 {
			break
		}
		if r.ID == Sum2 {
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
		return &sum
	}
	return nil
}

func (sc *ScoreCard) CalculateSums() {
	for _, col := range sc.Columns {
		if sc.Scores[Sum1][col.ID] == nil {
			sc.Scores[Sum1][col.ID] = sc.calcSum1(col.ID)
		}
		if sc.Scores[Sum2][col.ID] == nil {
			sc.Scores[Sum2][col.ID] = sc.calcSum2(col.ID)
		}
		if sc.Scores[Sum3][col.ID] == nil {
			sc.Scores[Sum3][col.ID] = sc.calcSum3(col.ID)
		}
	}
}

// check if all sum fields are filled (indicating completion)
func (sc *ScoreCard) IsComplete() bool {
	for _, col := range sc.Columns {
		if sc.Scores[Sum1][col.ID] == nil {
			return false
		}
		if sc.Scores[Sum2][col.ID] == nil {
			return false
		}
		if sc.Scores[Sum3][col.ID] == nil {
			return false
		}
	}
	return true
}

// total score across all sum fields
func (sc *ScoreCard) TotalScore() int {
	if !sc.IsComplete() {
		return 0
	}
	total := 0
	for _, col := range sc.Columns {
		total += *sc.Scores[Sum1][col.ID]
		total += *sc.Scores[Sum2][col.ID]
		total += *sc.Scores[Sum3][col.ID]
	}

	return total
}
