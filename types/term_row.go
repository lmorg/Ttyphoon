package types

import "strings"

type Row struct {
	Id      uint64
	Cells   []*Cell
	Hidden  Screen
	Phrase  *[]rune
	Source  *RowSource
	Block   *BlockMeta
	RowMeta RowMetaFlag
}

type RowMetaFlag uint16

const (
	META_ROW_NONE  RowMetaFlag = 0
	META_ROW_BEGIN RowMetaFlag = 1 << iota
	META_ROW_END
)

func (f RowMetaFlag) Is(flag RowMetaFlag) bool { return f&flag != 0 }
func (f *RowMetaFlag) Set(flag RowMetaFlag)    { *f |= flag }
func (f *RowMetaFlag) Unset(flag RowMetaFlag)  { *f &^= flag }

type RowSource struct {
	Host string
	Pwd  string
}

type BlockMeta struct {
	Query   []rune // typically command line
	ExitNum int
	Meta    BlockMetaFlag
}

type BlockMetaFlag uint16

const (
	META_BLOCK_NONE BlockMetaFlag = 0
	META_BLOCK_OK   BlockMetaFlag = 1 << iota
	META_BLOCK_ERROR
	META_BLOCK_AI
)

func (f BlockMetaFlag) Is(flag BlockMetaFlag) bool { return f&flag != 0 }
func (f *BlockMetaFlag) Set(flag BlockMetaFlag)    { *f |= flag }
func (f *BlockMetaFlag) Unset(flag BlockMetaFlag)  { *f &^= flag }

func (r *Row) String() string {
	slice := make([]rune, len(r.Cells))

	for i, cell := range r.Cells {
		slice[i] = cell.Rune()
	}

	return string(slice)
}

/*
	SCREEN
*/

type Screen []*Row

func (screen *Screen) String() string {
	slice := make([]string, len(*screen))
	for i, row := range *screen {
		slice[i] = row.String()
	}

	return strings.Join(slice, "\n")
}
