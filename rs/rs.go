package rs

import (
	"database/sql"
)

func GetResultSet(db *sql.DB, query string) ([]map[string]interface{}, error) {
	ret := []map[string]interface{}{}

	rows, eerr := db.Query(query)

	if eerr == nil {
		cols, cerr := rows.Columns()
		if cerr != nil {
			return nil, cerr
		}

		vals := make([]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {
			vals[i] = new(interface{})
		}

		for rows.Next() {
			scerr := rows.Scan(vals...)

			if scerr == nil {
				valmap := make(map[string]interface{})
				for i, col := range cols {
					valmap[col] = *(vals[i].(*interface{}))
				}

				ret = append(ret, valmap)
			} else {
				return nil, scerr
			}
		}
	} else {
		return nil, eerr
	}

	return ret, nil
}
