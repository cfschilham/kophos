package quadratic_sieve

/* homogenous system of linear equations with coefficients of GF(2) */

import (
	"fmt"
)

type Bit int

func (r Bit) Check() {
	if r != 0 && r != 1 {
		panic(fmt.Sprint("Invalid Value. Should be 0 or 1, but is ", r))
	}
}

func (r Bit) String() string {
	if r == 0 {
		return "0"
	} /* else { */

	return "1"
}

type Row struct {
	chunks      []uint64
	columnCount int
}

func NewRow(columnCount int) *Row {
	if columnCount < 0 {
		panic("column count has to be >= 0")
	}

	var ret Row
	ret.columnCount = columnCount
	ret.chunks = make([]uint64, ((columnCount-1)/64)+1)

	for i, _ := range ret.chunks {
		/* initialize to zero */
		ret.chunks[i] = 0x0000000000000000
	}

	return &ret
}

func (r Row) Column(index int) Bit {
	r.checkIndex(index)

	column, bit, exp := r.convertIndex(index)

	ret := Bit((r.chunks[column] & bit) >> exp)
	ret.Check()
	return ret
}

func (r *Row) SetColumn(index int, value Bit) {
	value.Check()
	r.checkIndex(index)

	column, bit, _ := r.convertIndex(index)

	if value == 0 {
		r.chunks[column] &= ^bit
	} else {
		r.chunks[column] |= bit
	}
}

func (r *Row) Set(other *Row) {
	r.checkSameSize(other)

	for i, k := range other.chunks {
		r.chunks[i] = k
	}
}

func (r *Row) Swap(other *Row) {

	r.checkSameSize(other)

	var tempChunk uint64
	for i, _ := range other.chunks {
		tempChunk = r.chunks[i]
		r.chunks[i] = other.chunks[i]
		other.chunks[i] = tempChunk
	}
}

func (r *Row) Xor(a, b *Row) {

	r.checkSameSize(a)
	r.checkSameSize(b)

	for i, _ := range a.chunks {
		r.chunks[i] = a.chunks[i] ^ b.chunks[i]
	}
}

func (r Row) IsZero() bool {
	for _, chunk := range r.chunks {
		if chunk != 0x0000000000000000 {
			return false
		}
	}
	return true
}

func (r Row) String() string {
	var ret string
	for i := r.columnCount - 1; i >= 0; i -= 1 {
		ret += fmt.Sprint(r.Column(i), " ")
	}
	return ret
}

func (r Row) Equals(other *Row) bool {

	r.checkSameSize(other)

	for i, chunk := range r.chunks {
		if chunk != other.chunks[i] {
			return false
		}
	}

	return true
}

func (r Row) checkIndex(index int) {
	if index < 0 || index >= r.columnCount {
		panic(fmt.Sprint("index out of bounds ", index, " !! [", 0, ",", r.columnCount, ")"))
	}
}

func (r Row) checkSameSize(other *Row) {
	if r.columnCount != other.columnCount {
		panic(fmt.Sprint("cannot perform r operation on two rows of differing size. columnCount ",
			r.columnCount, " != ", other.columnCount))
	}
}

func (r Row) convertIndex(index int) (column int, bit uint64, exp uint32) {
	return (len(r.chunks)*64 - 1 - index) / 64, 1 << (uint(index) % 64), uint32(index) % 64
}

type LinearSystem struct {
	rows                  []*Row
	rowCount, columnCount int
}

func NewLinearSystem(rows, columns int) *LinearSystem {

	if rows < 0 || columns < 0 {
		panic(fmt.Sprint("columnCount ", columns, " < 0 or rowCount ", rows, " < 0 "))
	}

	var ret LinearSystem
	ret.rowCount = rows
	ret.columnCount = columns
	ret.rows = make([]*Row, rows)

	for i, _ := range ret.rows {
		ret.rows[i] = NewRow(columns)
	}

	return &ret
}

func (l LinearSystem) Row(index int) *Row {
	l.checkRowIndex(index)
	return l.rows[index]
}

func (l *LinearSystem) SetRow(index int, row *Row) {
	l.checkRowIndex(index)
	l.rows[index].Set(row)
}

func (l *LinearSystem) Set(other *LinearSystem) {

	l.checkSameSize(other)

	for i, row := range other.rows {
		l.SetRow(i, row)
	}
}

func (l LinearSystem) EliminateEmptyRows() *LinearSystem {

	toBeKept := make(map[int]bool)

	emptyRow := NewRow(l.columnCount)

	for i, row := range l.rows {
		if row.Equals(emptyRow) == false {
			toBeKept[i] = true
		}
	}

	m := NewLinearSystem(len(toBeKept), l.columnCount)

	j := 0
	for i := 0; i < l.rowCount; i += 1 {
		if ok := toBeKept[i]; ok == true {
			m.SetRow(j, l.Row(i))
			j += 1
		}
	}

	return m
}

func (l *LinearSystem) GaussianElimination(other *LinearSystem) *LinearSystem {

	l.checkSameSize(other)

	l.Set(other)

	startingRow := 0

	for column := l.columnCount - 1; column >= 0; column -= 1 {

		var row int
		for row = startingRow; row < l.rowCount; row += 1 {
			if l.Row(row).Column(column) == 1 {
				l.Row(startingRow).Swap(l.Row(row))
				break
			}
		}

		if row == l.rowCount {
			/* no row has been found that has a bit at the wanted column,
			try again using the next column to the left */
			continue
		}

		for row = startingRow + 1; row < l.rowCount; row += 1 {
			if l.Row(row).Column(column) == 1 {
				l.Row(row).Xor(l.Row(row), l.Row(startingRow))
			}
		}

		startingRow += 1
	}

	return l
}

func (l *LinearSystem) MakeEmptyRows() [][]int {
	/* similar to gauss jordan */

	/* for each row keep the indices of the added rows */
	solution := NewLinearSystem(l.rowCount, l.rowCount)
	for i := 0; i < l.rowCount; i += 1 {
		solution.Row(i).SetColumn(i, 1)
	}

	startingRow := 0

	for column := l.columnCount - 1; column >= 0; column -= 1 {

		var row int
		for row = startingRow; row < l.rowCount; row += 1 {
			if l.Row(row).Column(column) == 1 {
				startingRow = row
				break
			}
		}

		if row == l.rowCount {
			/* no row has been found that has a bit at the wanted column,
			try again using the next column to the left */
			continue
		}

		for row = 0; row < l.rowCount; row += 1 {

			if row == startingRow {
				continue
			}

			if l.Row(row).Column(column) == 1 {
				l.Row(row).Xor(l.Row(row), l.Row(startingRow))
				solution.Row(row).Xor(solution.Row(row), solution.Row(startingRow))
			}
		}

		startingRow += 1
	}

	var ret [][]int

	for j := 0; j < l.rowCount; j += 1 {
		if l.Row(j).IsZero() == true {

			var solutionIndexSet []int

			for i := 0; i < solution.Row(j).columnCount; i += 1 {
				if solution.Row(j).Column(i) == 1 {
					solutionIndexSet = append(solutionIndexSet, i)
				}
			}

			ret = append(ret, solutionIndexSet)
		}
	}

	return ret
}

func (l LinearSystem) Transpose() *LinearSystem {
	m := NewLinearSystem(l.columnCount, l.rowCount)
	for j, row := range l.rows {
		for i := 0; i < row.columnCount; i += 1 {
			m.Row(i).SetColumn(j, row.Column(i))
		}
	}
	return m
}

func (l LinearSystem) String() string {
	var ret string
	for _, k := range l.rows {
		ret += fmt.Sprint(k)
		ret += "\n"
	}
	return ret
}

func (l LinearSystem) Equals(other *LinearSystem) bool {

	l.checkSameSize(other)

	for i, row := range l.rows {
		if row.Equals(other.Row(i)) == false {
			return false
		}
	}

	return true
}

/* *** private *** */
func (l LinearSystem) checkRowIndex(i int) {
	if i < 0 || i >= l.rowCount {
		panic(fmt.Sprint("invalid index ", i, " is not element of [0 ,", l.rowCount, ")"))
	}
}

func (l LinearSystem) checkSameSize(other *LinearSystem) {
	if l.rowCount != other.rowCount || l.columnCount != other.columnCount {
		panic(fmt.Sprint("cannot perform operation on two linear systems of differing size. columnCount ",
			l.rowCount, " != ", other.rowCount, " or ", l.columnCount, " != ", other.columnCount))
	}
}

/* *** helper *** ********************************************************** */
