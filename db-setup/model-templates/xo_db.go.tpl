// XODB is the common interface for database operations that can be used with
// types from schema '{{ schema .Schema }}'.
//
// This should work with database/sql.DB and database/sql.Tx.

var(
	ErrNotInDatabaseTable = errors.New("Table not in database table")
)

const(
	databaseTableErr = "Table \"%s\" not found in database_table.  Did you add it?"
)

func inclusionMap(field *structs.Field, listField string, listVal interface{}, values map[string]interface{}) (map[string]interface{}, error) {
	if field.IsExported() {
		structField := field.Tag("json")
		structField = strings.Split(structField, ",")[0]
		if listField == structField {
			switch field.Value().(type) {
			case string, *string, int64, *int64, float64, *float64, bool, *bool, uuid.UUID:
				switch field.Value().(type) {
				case int64:
					val := field.Value().(int64)
					values[structField] = strconv.FormatInt(val, confutil.IntBase)
				case *int64:
					val := field.Value().(*int64)
					values[structField] = strconv.FormatInt(*val, confutil.IntBase)
				case uuid.UUID:
					val := field.Value().(uuid.UUID)
					values[structField] = val.String()
				default:
					values[structField] = field.Value()
				}
			default:
				if listField == structField {
					iFieldVal, ok := listVal.(map[string]interface{})

					if !ok {
						return nil, fmt.Errorf("\"%s\" must contain map[string]interface{} for struct types", listField)
					}

					newVals := make(map[string]interface{}, 0)

					for _, f := range field.Fields() {
						for k, v := range iFieldVal {
							vals, err := inclusionMap(f, k, v, newVals)

							if err != nil {
								return nil, err
							}

							values[structField] = vals
						}
					}
				}
			}
		}
	}

	return values, nil
}

func exclusionMap(field *structs.Field, listField string, listVal interface{}) (map[string]interface{}, error) {
	var newVals map[string]interface{}

	if field.IsExported() {
		switch field.Value().(type) {
		case string, *string, int64, *int64, float64, *float64, bool, *bool:
			return nil, nil
		default:
			excludeMap, ok := listVal.(map[string]interface{})

			if !ok {
				return nil, fmt.Errorf("\"%s\" must contain map[string]interface{} for struct types", listField)
			}

			newVals = make(map[string]interface{}, 0)
			for _, f := range field.Fields() {
				innerStructField := f.Tag("json")
				innerStructField = strings.Split(innerStructField, ",")[0]

				if _, ok := excludeMap[innerStructField]; !ok {
					if f.IsExported() {
						//fmt.Printf("innerstruct include: %s\n", innerStructField)
						newVals[innerStructField] = f.Value()
					}
				} else {
					vals, err := exclusionMap(f, innerStructField, excludeMap[innerStructField])

					if err != nil {
						return nil, err
					}

					// values[innerStructField] = vals
					// fmt.Printf("inner vales: %v", values[innerStructField])

					if vals != nil {
						newVals[innerStructField] = vals
					}
				}
			}

			//fmt.Printf("inner returned map: %v\n\n", newVals)

			//fmt.Printf("map values: %v", newVals)
		}
	} else {
		return nil, nil
	}

	return newVals, nil
}

// XOLog provides the log func used by generated queries.
var XOLog = func(string, ...interface{}) { }

// ScannerValuer is the common interface for types that implement both the
// database/sql.Scanner and sql/driver.Valuer interfaces.
type ScannerValuer interface {
	sql.Scanner
	driver.Valuer
}

// StringSlice is a slice of strings.
type StringSlice []string

// quoteEscapeRegex is the regex to match escaped characters in a string.
var quoteEscapeRegex = regexp.MustCompile(`([^\\]([\\]{2})*)\\"`)

// Scan satisfies the sql.Scanner interface for StringSlice.
func (ss *StringSlice) Scan(contractorTracking interface{}) error {
	buf, ok := contractorTracking.([]byte)
	if !ok {
		return errors.New("invalid StringSlice")
	}

	// change quote escapes for csv parser
	str := quoteEscapeRegex.ReplaceAllString(string(buf), `$1""`)
	str = strings.Replace(str, `\\`, `\`, -1)

	// remove braces
	str = str[1:len(str)-1]

	// bail if only one
	if len(str) == 0 {
		*ss = StringSlice([]string{})
		return nil
	}

	// parse with csv reader
	cr := csv.NewReader(strings.NewReader(str))
	slice, err := cr.Read()
	if err != nil {
		fmt.Printf("exiting!: %v\n", err)
		return err
	}

	*ss = StringSlice(slice)

	return nil
}

// Value satisfies the driver.Valuer interface for StringSlice.
func (ss StringSlice) Value() (driver.Value, error) {
	v := make([]string, len(ss))
	for i, s := range ss {
		v[i] = `"` + strings.Replace(strings.Replace(s, `\`, `\\\`, -1), `"`, `\"`, -1) + `"`
	}
	return "{" + strings.Join(v, ",") + "}", nil
}

// Slice is a slice of ScannerValuers.
type Slice []ScannerValuer

