package table

// RowsToString type casts `[]Row` to `[][]string`.
func RowsToString(rows []Row) [][]string {
	result := make([][]string, len(rows))
	for i, row := range rows {
		result[i] = []string(row)
	}
	return result
}

// ColsToString type casts `[]Column` to `[]string`.
func ColsToString(cols []Column) []string {
	result := make([]string, len(cols))
	for i, col := range cols {
		result[i] = col.Title
	}
	return result
}
