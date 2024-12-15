package elementCsv

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/lmorg/mxtty/types"
)

type ElementCsv struct {
	renderer   types.Renderer
	size       types.XY
	headings   [][]rune // columns
	table      [][]rune // rendered rows
	top        []rune   // rendered headings
	width      []int    // columns
	boundaries []int32  // column lines
	isNumber   []bool   // columns

	//parameters parametersT

	name   string
	buf    []rune
	lines  int32
	notify types.Notification

	db   *sql.DB
	dbTx *sql.Tx

	filter       string
	orderByIndex int  // row
	orderDesc    bool // ASC or DESC

	renderOffset int32
	highlight    *types.XY
}

var arrowGlyph = map[bool]rune{
	false: '↑',
	true:  '↓',
}

const notifyLoading = "Loading CSV. Line %d..."

func New(renderer types.Renderer) *ElementCsv {
	el := &ElementCsv{renderer: renderer}

	el.notify = renderer.DisplaySticky(types.NOTIFY_INFO, fmt.Sprintf(notifyLoading, el.lines))

	err := el.createDb()
	if err != nil {
		panic(err)
	}

	return el
}

func (el *ElementCsv) Write(r rune) error {
	el.buf = append(el.buf, r)
	if r == '\n' {
		el.lines++
		el.notify.SetMessage(fmt.Sprintf(notifyLoading, el.lines))
	}
	return nil
}

func (el *ElementCsv) Generate(apc *types.ApcSlice) error {
	defer el.notify.Close()

	buf := bytes.NewBufferString(string(el.buf))
	r := csv.NewReader(buf)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %v", err)
	}

	if len(recs) < 2 {
		return fmt.Errorf("too few rows") // TODO: this shouldn't error
	}

	err = el.createTable(recs[0])
	if err != nil {
		return err
	}

	n := len(recs[0])

	el.headings = make([][]rune, n)
	for i := range recs[0] {
		el.headings[i] = []rune(recs[0][i])
	}

	// figure out if number
	el.isNumber = make([]bool, n)
	for col := 0; col < n && col < len(recs[1]); col++ {
		_, e := strconv.ParseFloat(recs[1][col], 64)
		el.isNumber[col] = e == nil // if no error, then it's probably a number
	}

	for row := 1; row < len(recs); row++ {
		if len(recs[row]) > n {
			recs[row][n-1] = strings.Join(recs[row][n-1:], " ")
			recs[row] = recs[row][:n]
		}
		err = el.insertRecords(recs[row])
		if err != nil {
			return err
		}
	}

	if el.dbTx.Commit() != nil {
		return fmt.Errorf("cannot commit sqlite3 transaction: %v", err)
	}

	el.size = *el.renderer.GetTermSize()
	if el.size.Y > 8 {
		el.size.Y -= 5
	}
	if el.size.Y > el.lines {
		el.size.Y = el.lines
	}

	err = el.runQuery()
	if err != nil {
		return err
	}

	return nil
}

func (el *ElementCsv) Size() *types.XY {
	return &el.size
}

func (el *ElementCsv) Rune(pos *types.XY) rune {
	pos.X -= el.renderOffset

	if pos.Y == 0 {
		if int(pos.X) >= len(el.top) {
			return ' '
		}
		return el.top[pos.X]
	}

	if int(pos.Y) > len(el.table) {
		return ' '
	}

	if int(pos.X) >= len(el.table[pos.Y-1]) {
		return ' '
	}

	return el.table[pos.Y-1][pos.X]
}

func (el *ElementCsv) Draw(size *types.XY, pos *types.XY) {
	pos.X += el.renderOffset

	cell := &types.Cell{Sgr: &types.Sgr{}}
	cell.Sgr.Reset()
	relPos := &types.XY{X: pos.X, Y: pos.Y}

	cell.Sgr.Bitwise |= types.SGR_INVERT
	for i := range el.top {
		cell.Char = el.top[i]
		el.renderer.PrintCell(cell, relPos)
		relPos.X++
	}

	switch el.orderByIndex {
	case 0:
		goto skipOrderGlyph

	case 1:
		relPos.X = pos.X + 0

	default:
		relPos.X = pos.X + el.boundaries[el.orderByIndex-2]
	}

	cell.Sgr.Bitwise |= types.SGR_BOLD
	cell.Char = arrowGlyph[el.orderDesc]
	el.renderer.PrintCell(cell, relPos)
	cell.Sgr.Bitwise ^= types.SGR_BOLD

skipOrderGlyph:

	relPos.Y++
	cell.Sgr.Bitwise ^= types.SGR_INVERT

	for y := int32(0); y < el.size.Y-1 && int(y) < len(el.table); y++ {
		relPos.X = 0
		for x := -el.renderOffset; x+el.renderOffset < el.size.X && int(x) < len(el.table[y]); x++ {
			cell.Char = el.table[y][x]
			el.renderer.PrintCell(cell, relPos)
			relPos.X++
		}
		relPos.Y++
	}

	el.renderer.DrawTable(pos, int32(len(el.table)), el.boundaries)

	if el.highlight != nil {
		var start, end int32

		for i := range el.boundaries {
			if el.highlight.X-el.renderOffset < el.boundaries[i] {
				if i != 0 {
					start = el.boundaries[i-1] + pos.X
					end = int32(el.width[i]) + 2
				} else {
					end = int32(el.width[i]) + 2 + el.renderOffset
				}
				break
			}
		}

		el.renderer.DrawHighlightRect(
			&types.XY{X: start, Y: el.highlight.Y + pos.Y},
			&types.XY{X: end, Y: 1},
		)
	}
}

func (el *ElementCsv) Close() {
	// clear memory (if required)
	el.db.Close()
}
