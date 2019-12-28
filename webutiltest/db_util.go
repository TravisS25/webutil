package webutiltest

import (
	"github.com/TravisS25/webutil/webutil"
	"github.com/pkg/errors"
)

func DBSetup(db webutil.QueryTransaction, bindVar int) func() error {
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
			logging.action_id = 1
		group by
			logging.primary_key_id,
			logging.primary_key_uuid,
			database_table.name
		order by
			min(logging.date_created) desc;
		`

		rower, err := db.Query(query)

		if err != nil {
			return errors.Wrap(err, "")
		}

		tx, _ := db.Begin()

		for rower.Next() {
			var logID, dbTableUUID, dbTableID interface{}
			var args []interface{}
			var dbTableName string

			err = rower.Scan(&logID, &dbTableID, &dbTableUUID, &dbTableName)

			if err != nil {
				return errors.Wrap(err, "")
			}

			logQuery := `delete from logging where id = ?;`
			logQuery, args, err = webutil.InQueryRebind(bindVar, query, logID)

			if err != nil {
				return errors.Wrap(err, "")
			}

			_, err = tx.Exec(logQuery, args...)

			if err != nil {
				return errors.Wrap(err, "")
			}

			tableQuery := `delete from ` + dbTableName + ` where id = ?;`

			if dbTableID != nil {
				tableQuery, args, err = webutil.InQueryRebind(bindVar, query, dbTableID)
			} else {
				tableQuery, args, err = webutil.InQueryRebind(bindVar, query, dbTableUUID)
			}

			if err != nil {
				return errors.Wrap(err, "")
			}

			_, err = tx.Exec(tableQuery, args...)

			if err != nil {
				return errors.Wrap(err, "")
			}
		}

		err = tx.Commit()

		if err != nil {
			return errors.Wrap(err, "")
		}

		_, err = db.Query(`delete from logging;`)

		if err != nil {
			return errors.Wrap(err, "")
		}

		return nil
	}
}
