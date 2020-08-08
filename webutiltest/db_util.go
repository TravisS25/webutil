package webutiltest

import (
	"strings"

	"github.com/TravisS25/webutil/webutil"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	foreignKeyViolation = "foreign_key_violation"
)

// DBSetup allows user to set up and tear down against a live
// database for each test without having to set up and tear down
// the entire database everytime
//
// This is accomplished by entering every action into a logging table
// and when the test is finished, delete all the records associated
// with the table
//
// Example
//
// func TestFoo(t *testing.T){
//		teardown := DBSetup(realDB, sqlx.DOLLAR)
//		defer teardown()
//
//		...
//		Test Code
// }
//
//
// func DBSetup(db webutil.QuerierTx, bindVar int) func() error {
// 	return func() error {
// 		query :=
// 			`
// 		select
// 			min(logging.id),
// 			logging.primary_key_id,
// 			logging.primary_key_uuid,
// 			database_table.name
// 		from
// 			logging
// 		join
// 			database_table on logging.database_table_id = database_table.id
// 		where
// 			logging.database_action_id = 1
// 		group by
// 			logging.primary_key_id,
// 			logging.primary_key_uuid,
// 			database_table.name
// 		order by
// 			min(logging.date_created) desc;
// 		`

// 		rower, err := db.Queryx(query)

// 		if err != nil {
// 			return errors.Wrap(err, "")
// 		}

// 		tx, _ := db.Beginx()

// 		for rower.Next() {
// 			var logID, dbTableUUID, dbTableID interface{}
// 			var args []interface{}
// 			var dbTableName string

// 			err = rower.Scan(&logID, &dbTableID, &dbTableUUID, &dbTableName)

// 			if err != nil {
// 				tx.Rollback()
// 				return errors.Wrap(err, "")
// 			}

// 			logQuery := `delete from logging where id = ?;`
// 			logQuery, args, err = webutil.InQueryRebind(bindVar, logQuery, logID)

// 			if err != nil {
// 				tx.Rollback()
// 				return errors.Wrap(err, "")
// 			}

// 			_, err = tx.Exec(logQuery, args...)

// 			if err != nil {
// 				tx.Rollback()
// 				return errors.Wrap(err, "")
// 			}

// 			tableQuery := `delete from ` + dbTableName + ` where id = ?;`

// 			if dbTableID != nil {
// 				tableQuery, args, err = webutil.InQueryRebind(bindVar, tableQuery, dbTableID)
// 			} else {
// 				tableQuery, args, err = webutil.InQueryRebind(bindVar, tableQuery, dbTableUUID)
// 			}

// 			if err != nil {
// 				tx.Rollback()
// 				return errors.Wrap(err, "")
// 			}

// 			_, err = tx.Exec(tableQuery, args...)

// 			if err != nil {
// 				if pgErr, isPGErr := err.(*pq.Error); isPGErr {
// 					if pgErr.Code.Name() == foreignKeyViolation {
// 						if err = foreignDeletion(tx, pgErr, dbTableName); err != nil {
// 							tx.Rollback()
// 							return errors.Wrap(err, "")
// 						}
// 					} else {
// 						tx.Rollback()
// 						return errors.Wrap(err, "")
// 					}
// 				} else {
// 					tx.Rollback()
// 					return errors.Wrap(err, "")
// 				}
// 			}
// 		}

// 		err = tx.Commit()

// 		if err != nil {
// 			return errors.Wrap(err, "")
// 		}

// 		_, err = db.Queryx(`delete from logging;`)

// 		if err != nil {
// 			return errors.Wrap(err, "")
// 		}

// 		return nil
// 	}
// }

// func foreignDeletion(tx *sqlx.Tx, err error, currentTable string) error {
// 	fmt.Printf("err under table: %s\n", currentTable)
// 	//tx.Rollback()

// 	var valuesIdx int
// 	var foreignColumnID string
// 	var tableToDelete string
// 	words := strings.Split(err.Error(), " ")

// 	for i, v := range words {
// 		if v == "values" {
// 			fmt.Printf("found values\n")
// 			valuesIdx = i
// 		}

// 		if strings.Contains(v, "[") && valuesIdx+1 == i {
// 			fmt.Printf("found id ref\n")
// 			foreignColumnID = v[1 : len(v)-1]
// 		}

// 		if strings.Contains(v, `"`) {
// 			tableToDelete = v[1 : len(v)-1]
// 		}
// 	}

// 	rows, err := tx.Queryx(
// 		`
// 		SELECT
// 			ccu.table_name,
// 			kcu.column_name
// 		FROM
// 			information_schema.table_constraints AS tc
// 		JOIN
// 			information_schema.key_column_usage AS kcu
// 		ON
// 			tc.constraint_name = kcu.constraint_name
// 		AND
// 			tc.table_schema = kcu.table_schema
// 		JOIN
// 			information_schema.constraint_column_usage AS ccu
// 		ON
// 			ccu.constraint_name = tc.constraint_name
// 		AND
// 			ccu.table_schema = tc.table_schema
// 		WHERE
// 			tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public' AND tc.table_name = $1
// 		GROUP by
// 			ccu.table_name, ccu.constraint_name, kcu.column_name
// 		ORDER BY
// 			ccu.table_name, kcu.column_name;
// 		`,
// 		tableToDelete,
// 	)

// 	if err != nil {
// 		fmt.Printf("error 1")
// 		fmt.Printf("table to delete: %s\n", tableToDelete)
// 		tx.Rollback()
// 		return errors.Wrap(err, "")
// 	}

// 	var foreignColumn string

// 	for rows.Next() {
// 		var tableName, columnName string

// 		if err = rows.Scan(&tableName, &columnName); err != nil {
// 			fmt.Printf("error 2")
// 			tx.Rollback()
// 			return errors.Wrap(err, "")
// 		}

// 		if tableName == currentTable {
// 			foreignColumn = columnName
// 			break
// 		}
// 	}

// 	if foreignColumn == "" {
// 		tx.Rollback()
// 		return errors.New("could not find foreign column")
// 	}

// 	query := "delete from " + tableToDelete + " where " + foreignColumn + " = " + foreignColumnID

// 	fmt.Printf("delete query: %s\n", query)

// 	if _, err = tx.Exec(query); err != nil {
// 		if pgErr, isPGErr := err.(*pq.Error); isPGErr {
// 			if pgErr.Code.Name() == foreignKeyViolation {
// 				if err = foreignDeletion(tx, pgErr, tableToDelete); err != nil {
// 					tx.Rollback()
// 					return errors.Wrap(err, "")
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

func DBSetup(db webutil.QuerierExec, bindVar int) func() error {
	return func() error {
		query :=
			`
		select 
			min(logging.id),
			logging.primary_key_id,
			logging.primary_key_uuid,
			database_table.name
		from 
			logging
		join
			database_table on logging.database_table_id = database_table.id
		where
			logging.database_action_id = 1
		group by
			logging.primary_key_id,
			logging.primary_key_uuid,
			database_table.name
		order by
			min(logging.date_created) desc;
		`

		rower, err := db.Queryx(query)

		if err != nil {
			return errors.Wrap(err, "")
		}

		//tx, _ := db.Beginx()

		for rower.Next() {
			var logID, dbTableUUID, dbTableID interface{}
			var args []interface{}
			var dbTableName string

			err = rower.Scan(&logID, &dbTableID, &dbTableUUID, &dbTableName)

			if err != nil {
				//tx.Rollback()
				return errors.Wrap(err, "")
			}

			logQuery := `delete from logging where id = ?;`
			logQuery, args, err = webutil.InQueryRebind(bindVar, logQuery, logID)

			if err != nil {
				//tx.Rollback()
				return errors.Wrap(err, "")
			}

			_, err = db.Exec(logQuery, args...)

			if err != nil {
				//tx.Rollback()
				return errors.Wrap(err, "")
			}

			tableQuery := `delete from ` + dbTableName + ` where id = ?;`

			if dbTableID != nil {
				tableQuery, args, err = webutil.InQueryRebind(bindVar, tableQuery, dbTableID)
			} else {
				tableQuery, args, err = webutil.InQueryRebind(bindVar, tableQuery, dbTableUUID)
			}

			if err != nil {
				//tx.Rollback()
				return errors.Wrap(err, "")
			}

			_, err = db.Exec(tableQuery, args...)

			if err != nil {
				if pgErr, isPGErr := err.(*pq.Error); isPGErr {
					if pgErr.Code.Name() == foreignKeyViolation {
						if err = foreignDeletion(db, pgErr, dbTableName); err != nil {
							//tx.Rollback()
							return errors.Wrap(err, "")
						}
					} else {
						//tx.Rollback()
						return errors.Wrap(err, "")
					}
				} else {
					//tx.Rollback()
					return errors.Wrap(err, "")
				}
			}
		}

		// err = tx.Commit()

		// if err != nil {
		// 	return errors.Wrap(err, "")
		// }

		_, err = db.Queryx(`delete from logging;`)

		if err != nil {
			return errors.Wrap(err, "")
		}

		return nil
	}
}

func foreignDeletion(db webutil.QuerierExec, err error, currentTable string) error {
	// fmt.Printf("err under table: %s\n", currentTable)
	// tx.Rollback()

	var valuesIdx int
	var foreignColumnID string
	var tableToDelete string
	words := strings.Split(err.Error(), " ")

	for i, v := range words {
		if v == "values" {
			//fmt.Printf("found values\n")
			valuesIdx = i
		}

		if strings.Contains(v, "[") && valuesIdx+1 == i {
			//fmt.Printf("found id ref\n")
			foreignColumnID = v[1 : len(v)-1]
		}

		if strings.Contains(v, `"`) {
			tableToDelete = v[1 : len(v)-1]
		}
	}

	rows, err := db.Queryx(
		`
		SELECT 
			ccu.table_name,
			kcu.column_name
		FROM 
			information_schema.table_constraints AS tc 
		JOIN 
			information_schema.key_column_usage AS kcu
		ON 
			tc.constraint_name = kcu.constraint_name
		AND 
			tc.table_schema = kcu.table_schema
		JOIN 
			information_schema.constraint_column_usage AS ccu
		ON 
			ccu.constraint_name = tc.constraint_name
		AND 
			ccu.table_schema = tc.table_schema
		WHERE 
			tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public' AND tc.table_name = $1
		GROUP by
			ccu.table_name, ccu.constraint_name, kcu.column_name
		ORDER BY
			ccu.table_name, kcu.column_name;
		`,
		tableToDelete,
	)

	if err != nil {
		return errors.Wrap(err, "")
	}

	var foreignColumn string

	for rows.Next() {
		var tableName, columnName string

		if err = rows.Scan(&tableName, &columnName); err != nil {
			return errors.Wrap(err, "")
		}

		if tableName == currentTable {
			foreignColumn = columnName
			break
		}
	}

	if foreignColumn == "" {
		return errors.New("could not find foreign column")
	}

	query := "delete from " + tableToDelete + " where " + foreignColumn + " = " + foreignColumnID

	if _, err = db.Exec(query); err != nil {
		if pgErr, isPGErr := err.(*pq.Error); isPGErr {
			if pgErr.Code.Name() == foreignKeyViolation {
				if err = foreignDeletion(db, pgErr, tableToDelete); err != nil {
					// /tx.Rollback()
					return errors.Wrap(err, "")
				}
			}
		}
	}

	return nil
}
