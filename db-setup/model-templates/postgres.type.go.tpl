{{- $short := (shortname .Name "err" "res" "sqlstr" "db" "XOLog") -}}
{{- $table := (schema .Schema .Table.TableName) -}}
{{- if .Comment -}}
// {{ .Comment }}
{{- else -}}
// {{ .Name }} represents a row from '{{ $table }}'.
{{- end }}
type {{ .Name }} struct {
{{- range .Fields }}
	{{ .Name }} {{ convertType .Type }} `json:"{{ ignoreJSONFields $.Table.TableName .Col.ColumnName .Type }}" db:"{{ .Col.ColumnName }}"` // {{ .Col.ColumnName }}
{{- end }}

{{- range .ForeignKeys }}
	{{ refColumn .ColumnName true }} *{{ refTable .RefTableName }} `json:"{{ refColumnJSON .ColumnName }}" db:"{{ refColumn .ColumnName false }}"`
{{- end }}

excludeJSONFields map[string]interface{}
includeJSONFields map[string]interface{}
}

func Query{{ .Name }}(db webutil.SqlxDB, bindVar int, query string, args ...interface{}) (*{{ .Name }}, error) {
	var dest {{.Name }}
	var err error

	if query, args, err = webutil.InQueryRebind(bindVar, query, args...); err != nil{
		return nil, err
	}

	err = db.Get(&dest, query, args...)
	return &dest, err
}

func Query{{ convertName .Name }}s(db webutil.SqlxDB, bindVar int, query string, args ...interface{}) ([]*{{ .Name }}, error) {
	var dest []*{{.Name }}
	var err error

	if query, args, err = webutil.InQueryRebind(bindVar, query, args...); err != nil{
		return nil, err
	}

	err = db.Select(&dest, query, args...)
	return dest, err
}

func ({{ $short }} *{{ .Name }}) MarshalJSON() ([]byte, error) {
	var value interface{}
	var err error

	if len({{ $short }}.excludeJSONFields) != 0 {
		values := make(map[string]interface{}, 0)
		fields := structs.Fields({{ $short }})

		for _, f := range fields {
			structField := f.Tag("json")
			structField = strings.Split(structField, ",")[0]

			if _, ok := {{ $short }}.excludeJSONFields[structField]; !ok {
				if f.IsExported() {
					values[structField] = f.Value()
				}
			} else {
				vals, err := exclusionMap(f, structField, {{ $short }}.excludeJSONFields[structField])

				if err != nil {
					return nil, err
				}

				//fmt.Printf("outter returned map: %v\n\n", vals)

				if len(vals) != 0 {
					values[structField] = vals
				}
			}
		}

		//fmt.Printf("final values: %v\n\n", values)

		value = values
	} else if len({{ $short }}.includeJSONFields) != 0 {
		values := make(map[string]interface{}, 0)
		fields := structs.Fields({{ $short }})

		for _, f := range fields {
			for k, v := range {{ $short }}.includeJSONFields {
				_, err = inclusionMap(f, k, v, values)

				if err != nil {
					return nil, err
				}
			}
		}

		value = values
	} else {
		value = *{{ $short }}
	}

	return json.Marshal(&value)
}

func ({{ $short }} *{{ .Name }}) SetExclusionJSONFields(fields map[string]interface{}) {
	{{ $short }}.excludeJSONFields = fields
}

func ({{ $short }} *{{ .Name }}) SetInclusionJSONFields(fields map[string]interface{}) {
	{{ $short }}.includeJSONFields = fields
}

// Insert inserts the {{ .Name }} to the database.
func ({{ $short }} *{{ .Name }}) Insert(db webutil.QuerierExec) error {
	var err error

{{ if .Table.ManualPk }}
	// sql insert query, primary key must be provided
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames .Fields }}` +
		`) VALUES (` +
		`{{ colvals .Fields }}` +
		`)`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields $short }})
	_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short }})
	
	if err != nil {
		return err
	}
{{ else }}
	// sql insert query, primary key provided by sequence
	const sqlstr = `INSERT INTO {{ $table }} (` +
		`{{ colnames .Fields .PrimaryKey.Name }}` +
		`) VALUES (` +
		`{{ colvals .Fields .PrimaryKey.Name }}` +
		`) RETURNING {{ colname .PrimaryKey.Col }}`

	// run query
	XOLog(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }})
	err = db.QueryRow(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}).Scan(&{{ $short }}.{{ .PrimaryKey.Name }})
	if err != nil {
		return err
	}
{{ end }}

	return nil
}

// Insert inserts the {{ .Name }} to the database along with adding to logging table.
func ({{ $short }} *{{ .Name }}) InsertWithLog(db webutil.Entity, bindVar int, request *http.Request) error {
	var err error
	err = {{ $short }}.Insert(db)

	if err != nil{
		return err
	}

	err = {{ $short }}.insertLog(db, request, 1, bindVar)

	if err == ErrNotInDatabaseTable{
		message := fmt.Sprint(databaseTableErr, "{{ .Name }}")
		return errors.New(message)
	}

	return nil
}

{{ if ne (fieldnamesmulti .Fields $short .PrimaryKeyFields) "" }}
	// Update updates the {{ .Name }} in the database.
	func ({{ $short }} *{{ .Name }}) Update(db webutil.QuerierExec) error {
		var err error

		{{ if gt ( len .PrimaryKeyFields ) 1 }}
			// sql query with composite primary key
			const sqlstr = `UPDATE {{ $table }} SET (` +
				`{{ colnamesmulti .Fields .PrimaryKeyFields }}` +
				`) = ( ` +
				`{{ colvalsmulti .Fields .PrimaryKeyFields }}` +
				`) WHERE {{ colnamesquerymulti .PrimaryKeyFields " AND " (getstartcount .Fields .PrimaryKeyFields) nil }}`

			// run query
			XOLog(sqlstr, {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})
			_, err = db.Exec(sqlstr, {{ fieldnamesmulti .Fields $short .PrimaryKeyFields }}, {{ fieldnames .PrimaryKeyFields $short}})
		return err
		{{- else }}
			{{- if .PrimaryKey }}
				// sql query
				const sqlstr = `UPDATE {{ $table }} SET (` +
					`{{ colnames .Fields .PrimaryKey.Name }}` +
					`) = ( ` +
					`{{ colvals .Fields .PrimaryKey.Name }}` +
					`) WHERE {{ colname .PrimaryKey.Col }} = ${{ colcount .Fields .PrimaryKey.Name }}`

				// run query
				_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short .PrimaryKey.Name }}, {{ $short }}.{{ .PrimaryKey.Name }})
				return err
			{{- else }}
				// sql query
				const sqlstr = `UPDATE {{ $table }} SET (` +
					`{{ colnames .Fields }}` +
					`) = ( ` +
					`{{ colvals .Fields }}` +
					`) WHERE id = ${{ colcount .Fields }}`

				// run query
				_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short }}, {{ $short }}.ID.String())
				return err
			{{- end }}
		{{- end }}
	}


	func ({{ $short }} *{{ .Name }}) UpdateWithLog(db webutil.Entity, bindVar int, request *http.Request) error {
		var err error
		canInsertLog := {{ $short }}.canInsertLog(db, bindVar)
		err = {{ $short }}.Update(db)

		if err != nil{
			return err
		}

		if canInsertLog{
			err = {{ $short }}.insertLog(db, request, 2, bindVar)

			if err == ErrNotInDatabaseTable{
				message := fmt.Sprint(databaseTableErr, "{{ .Name }}")
				return errors.New(message)
			}
		}
		return nil
	}


	// Upsert performs an upsert for {{ .Name }}.
	//
	// NOTE: PostgreSQL 9.5+ only
	func ({{ $short }} *{{ .Name }}) Upsert(db webutil.QuerierExec) error {
		var err error

		// sql query
		const sqlstr = `INSERT INTO {{ $table }} (` +
			`{{ colnames .Fields }}` +
			`) VALUES (` +
			`{{ colvals .Fields }}` +
			`) ON CONFLICT ({{ colnames .PrimaryKeyFields }}) DO UPDATE SET (` +
			`{{ colnames .Fields }}` +
			`) = (` +
			`{{ colprefixnames .Fields "EXCLUDED" }}` +
			`)`

		// run query
		XOLog(sqlstr, {{ fieldnames .Fields $short }})
		_, err = db.Exec(sqlstr, {{ fieldnames .Fields $short }})
		if err != nil {
			return err
		}

		return nil
	}

{{ else }}
	// Update statements omitted due to lack of fields other than primary key
{{ end }}

// Delete deletes the {{ .Name }} from the database.
func ({{ $short }} *{{ .Name }}) Delete(db webutil.QuerierExec, bindVar int) error {
	var err error

	{{ if gt ( len .PrimaryKeyFields ) 1 }}
		// sql query with composite primary key
		const sqlstr = `DELETE FROM {{ $table }}  WHERE {{ colnamesquery .PrimaryKeyFields " AND " }}`

		// run query
		XOLog(sqlstr, {{ fieldnames .PrimaryKeyFields $short }})
		_, err = db.Exec(sqlstr, {{ fieldnames .PrimaryKeyFields $short }})
		if err != nil {
			return err
		}
	{{- else }}
		{{- if .PrimaryKey }}
			// sql query
			sqlstr := `DELETE FROM {{ $table }} WHERE {{ colname .PrimaryKey.Col }} = ?`

			if sqlstr, _, err = webutil.InQueryRebind(bindVar, sqlstr); err != nil{
				return err
			}

			// run query
			XOLog(sqlstr, {{ $short }}.{{ .PrimaryKey.Name }})
			_, err = db.Exec(sqlstr, {{ $short }}.{{ .PrimaryKey.Name }})
			if err != nil {
				return err
			}
		{{- else }}
			// sql query
			sqlstr := `DELETE FROM {{ $table }} WHERE id = ?`

			if sqlstr, _, err = webutil.InQueryRebind(bindVar, sqlstr); err != nil{
				return err
			}

			// run query
			_, err = db.Exec(sqlstr, {{ $short }}.ID.String())
			if err != nil {
				return err
			}
		{{- end }}
	{{- end }}

	return nil
}


//  Delete deletes the {{ .Name }} from the database while logging it.
func ({{ $short }} *{{ .Name }}) DeleteWithLog(db webutil.Entity, bindVar int, request *http.Request) error {
	var err error
	err = {{ $short }}.Delete(db, bindVar)

	if err != nil{
		return err
	}

	err = {{ $short }}.insertLog(db, request, 3, bindVar)

	if err == ErrNotInDatabaseTable{
		message := fmt.Sprint(databaseTableErr, "{{ .Name }}")
		return errors.New(message)
	}
	return nil
}

func ({{ $short }} *{{ .Name }}) canInsertLog(db webutil.Entity, bindVar int) bool{
	{{- if .PrimaryKey }}
		prev, err := Query{{ .Name }}(db, bindVar, `select * from {{ $table }} where {{ .PrimaryKey.Col.ColumnName }} = ?`, {{ $short }}.{{ .PrimaryKey.Name }})
	{{- else }}
		prev, err := Query{{ .Name }}(db, bindVar, `select * from {{ $table }} where id = ?`, {{ $short }}.ID.String())
	{{- end }}

	if err == nil{
		if cmp.Equal({{ $short }}, prev, cmpopts.IgnoreTypes(map[string]interface{}{})){
			return false
		}
	} else{
		return false
	}

	return true
}

func ({{ $short }} *{{ .Name }}) insertLog(db webutil.Entity, request *http.Request, dbAction, bindVar int) error {
	var args []interface{}

	rowBytes, err := json.Marshal(&{{ $short }})

	if err != nil{
		rowBytes = nil
	}

	var userID *int64

	if request.Context().Value(webutil.UserCtxKey) != nil {
		userBytes := request.Context().Value(webutil.UserCtxKey).([]byte)
		var user *UserProfile
		json.Unmarshal(userBytes, &user)

		userID = &user.ID
	}

	var tableID interface{} 
	tableName := "{{ .Table.TableName }}"

	query := `select id from database_table where name = ?;`

	if query, args, err = webutil.InQueryRebind(bindVar, query, tableName); err != nil{
		return err
	}

	table := db.QueryRow(query, args...)
	err = table.Scan(&tableID)

	if err != nil{
		fmt.Printf("Could not find table %s in database_table.  Did you add it?\n", tableName)
		return err
	}

	var areaID interface{}
	newStruct := structs.New({{ $short }})
	
	if field, ok := newStruct.FieldOk("AreaID"); ok{
		areaID = field.Value()
	}

	var id interface{}
	var uid interface{}

	{{- if .PrimaryKey }}
		{{- if idIsInt64 .PrimaryKey.Type }}
			id = {{ $short }}.{{ .PrimaryKey.Name }}
		{{- else }}
			uid = {{ $short }}.{{ .PrimaryKey.Name }}
		{{- end }}
	{{- else }}
		if {{ $short }}.ID.String() == ""{
			id, _ = uuid.NewV4()
		} 
	{{- end }}

	_, err = db.Exec(
		`
		insert into logging (date_created, data, primary_key_id, primary_key_uuid, been_viewed, http_method_id, database_table_id, user_profile_id)
		values($1, $2, $3, $4, $5, $6, $7, $8)
		`, 
		time.Now().UTC().Format(webutil.DateTimeMilliLayout),
		rowBytes,
		id,
		uid,
		false,
		dbAction,
		tableID,
		userID,
		areaID,
	)
	
	return nil
}