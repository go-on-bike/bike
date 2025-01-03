package assert

import "database/sql"

func DBNil(db *sql.DB) {
	condition := db == nil
	errMsg := "DB is not nil"
	assert(condition, errMsg)
}
