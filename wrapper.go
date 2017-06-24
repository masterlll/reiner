package reiner

import (
	"database/sql"
	"fmt"
	"strings"

	// The MySQL driver.
	_ "github.com/go-sql-driver/mysql"
)

type tableName string

type function struct {
	query  string
	values []interface{}
}
type condition struct {
	args      []interface{}
	connector string
}

type order struct {
	column string
	args   []interface{}
}

type join struct {
	table      string
	typ        string
	condition  string
	conditions []condition
}

// Wrapper represents a database connection.
type Wrapper struct {
	db                 *DB
	isSubQuery         bool
	query              string
	alias              string
	tableName          []string
	conditions         []condition
	havingConditions   []condition
	queryOptions       []string
	destination        interface{}
	joins              map[tableName]join
	params             []interface{}
	onDuplicateColumns []string
	lastInsertIDColumn string
	limit              []int
	orders             []order

	//
	Timestamp *Timestamp
	// Count is the count of the results, or the affected rows.
	Count int
	//
	TotalCount int
	//
	PageLimit int
	//
	TotalPage int
	// LasyQuery is last executed query.
	LastQuery string
	//
	LastInsertID int
	//
	LastInsertIDs []int
	//
	LastRows *sql.Rows
	//
	LastRow *sql.Row
}

// New creates a new database connection which provides the MySQL wrapper functions.
// The first data source name is for the master, the rest are for the slaves, which is used for the read/write split.
//     .New("root:root@/master", []string{"root:root@/slave", "root:root@/slave2"})
// Check https://dev.mysql.com/doc/refman/5.7/en/replication-solutions-scaleout.html for more information.
func newWrapper(db *DB) *Wrapper {
	return &Wrapper{db: db}
}

func (w *Wrapper) clean() {
	w.tableName = []string{}
	w.params = []interface{}{}
	w.orders = []order{}
	w.conditions = []condition{}
	w.havingConditions = []condition{}
	w.limit = []int{}
	w.query = ""
}

func (w *Wrapper) buildPair(data interface{}) {
	//switch v := data.(type) {
	//case *Wrapper:
	//}
}

func (w *Wrapper) bindParams(data interface{}) (query string) {
	switch d := data.(type) {
	case []interface{}:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	case []int:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	case []string:
		for _, v := range d {
			query += fmt.Sprintf("%s, ", w.bindParam(v))
		}
	}
	query = trim(query)
	return
}

func (w *Wrapper) bindParam(data interface{}) (param string) {
	switch v := data.(type) {
	case *Wrapper:
		if len(v.params) > 0 {
			w.params = append(w.params, v.params...)
		}
	case function:
		if len(v.values) > 0 {
			w.params = append(w.params, v.values...)
		}
	default:
		w.params = append(w.params, data)
	}
	param = w.paramToQuery(data)
	return
}

func (w *Wrapper) paramToQuery(data interface{}) (param string) {
	switch v := data.(type) {
	case *Wrapper:
		param = fmt.Sprintf("(%s)", v.query)
	case function:
		param = v.query
	case nil:
		param = "NULL"
	default:
		param = "?"
	}
	return
}

func (w *Wrapper) buildDuplicate() (query string) {
	if len(w.onDuplicateColumns) == 0 {
		return
	}
	query += "ON DUPLICATE KEY UPDATE "
	if w.lastInsertIDColumn != "" {
		query += fmt.Sprintf("%s=LAST_INSERT_ID(%s), ", w.lastInsertIDColumn, w.lastInsertIDColumn)
	}
	for _, v := range w.onDuplicateColumns {
		query += fmt.Sprintf("%s = VALUE(%s), ", v, v)
	}
	query = trim(query)
	return
}

func (w *Wrapper) buildInsert(operator string, data interface{}) (query string) {
	var columns, values, options string
	if len(w.queryOptions) > 0 {
		options = fmt.Sprintf("%s ", strings.Join(w.queryOptions, ", "))
	}

	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			columns += fmt.Sprintf("%s, ", column)
			values += fmt.Sprintf("%s, ", w.bindParam(value))
		}
		values = fmt.Sprintf("(%s)", trim(values))

	case []map[string]interface{}:
		for index, single := range realData {
			var currentValues string
			for column, value := range single {
				// Get the column names from the first data set only.
				if index == 0 {
					columns += fmt.Sprintf("%s, ", column)
				}
				currentValues += fmt.Sprintf("%s, ", w.bindParam(value))
			}
			values += fmt.Sprintf("(%s), ", trim(currentValues))
		}
		values = trim(values)
	}
	columns = trim(columns)
	query = fmt.Sprintf("%s %sINTO %s (%s) VALUES %s ", operator, options, w.tableName[0], columns, values)
	return
}

func (w *Wrapper) Table(tableName ...string) *Wrapper {
	w.tableName = tableName
	return w
}

func (w *Wrapper) Insert(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	w.query += w.buildDuplicate()
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query += w.buildLimit()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) InsertMulti(data interface{}) (err error) {
	w.query = w.buildInsert("INSERT", data)
	w.query += w.buildDuplicate()
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) Replace(data interface{}) (err error) {
	w.query = w.buildInsert("REPLACE", data)
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildLimit()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) Func(query string, data ...interface{}) function {
	return function{
		query:  query,
		values: data,
	}
}

func (w *Wrapper) Now(formats ...string) function {
	query := "NOW() "
	unitMap := map[string]string{
		"Y": "YEAR",
		"M": "MONTH",
		"D": "DAY",
		"W": "WEEK",
		"h": "HOUR",
		"m": "MINUTE",
		"s": "SECOND",
	}
	for _, v := range formats {
		operator := string(v[0])
		interval := v[1 : len(v)-1]
		unit := string(v[len(v)-1])
		query += fmt.Sprintf("%s INTERVAL %s %s ", operator, interval, unitMap[unit])
	}
	return function{
		query: strings.TrimSpace(query),
	}
}

func (w *Wrapper) OnDuplicate(columns []string, lastInsertID ...string) *Wrapper {
	w.onDuplicateColumns = columns
	if len(lastInsertID) != 0 {
		w.lastInsertIDColumn = lastInsertID[0]
	}
	return w
}

func (w *Wrapper) buildUpdate(data interface{}) (query string) {
	var set string
	query = fmt.Sprintf("UPDATE %s SET ", w.tableName[0])
	switch realData := data.(type) {
	case map[string]interface{}:
		for column, value := range realData {
			set += fmt.Sprintf("%s = %s, ", column, w.bindParam(value))
		}
	}
	query += fmt.Sprintf("%s ", trim(set))
	return
}

func (w *Wrapper) buildLimit() (query string) {
	switch len(w.limit) {
	case 0:
		return
	case 1:
		query = fmt.Sprintf("LIMIT %d ", w.limit[0])
	case 2:
		query = fmt.Sprintf("LIMIT %d, %d ", w.limit[0], w.limit[1])
	}
	return
}

func (w *Wrapper) Update(data interface{}) (err error) {
	w.query = w.buildUpdate(data)
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query += w.buildLimit()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) Limit(count int, to ...int) *Wrapper {
	if len(to) == 0 {
		w.limit = []int{count}
	} else {
		w.limit = []int{count, to[0]}
	}
	return w
}

func (w *Wrapper) buildSelect(columns ...string) (query string) {
	if len(columns) == 0 {
		query = fmt.Sprintf("SELECT * FROM %s ", w.tableName[0])
	} else {
		query = fmt.Sprintf("SELECT %s FROM %s ", strings.Join(columns, ", "), w.tableName[0])
	}
	return
}

func (w *Wrapper) Get(columns ...string) (err error) {
	w.query = w.buildSelect(columns...)
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query += w.buildLimit()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) GetOne(columns ...string) (err error) {
	err = w.Limit(1).Get(columns...)
	return
}

func (w *Wrapper) GetValue(column string) (err error) {
	err = w.Get(fmt.Sprintf("%s AS Value", column))
	return
}

func (w *Wrapper) Paginate(pageCount int, columns ...string) (err error) {
	err = w.Limit(w.PageLimit*(pageCount-1), w.PageLimit).Get(columns...)
	w.TotalPage = w.TotalCount / w.PageLimit
	return
}

func (w *Wrapper) RawQuery(query string, values ...interface{}) (err error) {
	w.query = query
	w.LastQuery = w.query
	w.bindParams(values)
	return
}

func (w *Wrapper) RawQueryOne(query string, values ...interface{}) (err error) {
	err = w.RawQuery(query, values...)
	return
}

func (w *Wrapper) RawQueryValue(query string, values ...interface{}) (err error) {
	return
}

func (w *Wrapper) buildWhere(typ string) (query string) {
	var conditions []condition
	if typ == "HAVING" {
		conditions = w.havingConditions
		if len(conditions) == 0 {
			return
		}
		query = "HAVING "
	} else {
		conditions = w.conditions
		if len(conditions) == 0 {
			return
		}
		query = "WHERE "
	}

	if len(conditions) == 0 {
		return
	}

	for i, v := range conditions {
		if i != 0 {
			query += fmt.Sprintf("%s ", v.connector)
		}

		var typ string
		switch q := v.args[0].(type) {
		case string:
			if strings.Contains(q, "?") || strings.Contains(q, "(") || len(v.args) == 1 {
				typ = "Query"
			} else {
				typ = "Column"
			}
		case *Wrapper:
			typ = "SubQuery"
		}

		switch len(v.args) {
		// .Where("Column = Column")
		case 1:
			query += fmt.Sprintf("%s ", v.args[0].(string))
		// .Where("Column = ?", "Value")
		// .Where("Column", "Value")
		// .Where(subQuery, "EXISTS")
		case 2:
			switch typ {
			case "Query":
				query += fmt.Sprintf("%s ", v.args[0].(string))
				w.bindParam(v.args[1])
			case "Column":
				query += fmt.Sprintf("%s = %s ", v.args[0].(string), w.bindParam(v.args[1]))
			case "SubQuery":
				query += fmt.Sprintf("%s %s ", v.args[1].(string), w.bindParam(v.args[0]))
			}
		// .Where("Column", ">", "Value")
		// .Where("Column", "IN", subQuery)
		// .Where("Column", "IS", nil)
		case 3:
			if v.args[1].(string) == "IN" || v.args[1].(string) == "NOT IN" {
				query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2]))
			} else {
				query += fmt.Sprintf("%s %s %s ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2]))
			}

		// .Where("(Column = ? OR Column = SHA(?))", "Value", "Value")
		// .Where("Column", "BETWEEN", 1, 20)
		default:
			if typ == "Query" {
				query += fmt.Sprintf("%s ", v.args[0].(string))
				w.bindParams(v.args[1:])
			} else {
				switch v.args[1].(string) {
				case "BETWEEN", "NOT BETWEEN":
					query += fmt.Sprintf("%s %s %s AND %s ", v.args[0].(string), v.args[1].(string), w.bindParam(v.args[2]), w.bindParam(v.args[3]))
				case "IN", "NOT IN":
					query += fmt.Sprintf("%s %s (%s) ", v.args[0].(string), v.args[1].(string), w.bindParams(v.args[2:]))
				}
			}
		}
	}
	return
}

func (w *Wrapper) saveCondition(typ, connector string, args ...interface{}) {
	var c condition
	c.connector = connector
	c.args = args
	if typ == "HAVING" {
		w.havingConditions = append(w.havingConditions, c)
	} else {
		w.conditions = append(w.conditions, c)
	}
}

func (w *Wrapper) Where(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "AND", args...)
	return w
}

func (w *Wrapper) OrWhere(args ...interface{}) *Wrapper {
	w.saveCondition("WHERE", "OR", args...)
	return w
}

func (w *Wrapper) Having(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "AND", args...)
	return w
}

func (w *Wrapper) OrHaving(args ...interface{}) *Wrapper {
	w.saveCondition("HAVING", "OR", args...)
	return w
}

func (w *Wrapper) buildDelete(tableNames ...string) (query string) {
	query += fmt.Sprintf("DELETE FROM %s ", strings.Join(tableNames, ", "))
	return
}

func (w *Wrapper) Delete() (err error) {
	w.query = w.buildDelete(w.tableName...)
	w.query += w.buildWhere("WHERE")
	w.query += w.buildWhere("HAVING")
	w.query += w.buildOrderBy()
	w.query += w.buildLimit()
	w.query = strings.TrimSpace(w.query)
	w.LastQuery = w.query
	w.clean()
	return
}

func (w *Wrapper) buildOrderBy() (query string) {
	if len(w.orders) == 0 {
		return
	}
	query += "ORDER BY "
	for _, v := range w.orders {
		if len(v.args) == 1 {
			query += fmt.Sprintf("%s %s, ", v.column, v.args[0])
		} else if len(v.args) > 1 {
			query += fmt.Sprintf("FIELD (%s, %s) %s, ", v.column, w.bindParams(v.args[1:]), v.args[0])
		} else {
			query += fmt.Sprintf("%s, ", v.column)
		}
	}
	query = trim(query)
	return
}

func (w *Wrapper) OrderBy(column string, args ...interface{}) *Wrapper {
	w.orders = append(w.orders, order{
		column: column,
		args:   args,
	})
	return w
}

func (w *Wrapper) GroupBy(column string) *Wrapper {
	return w
}

func (w *Wrapper) LeftJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) RightJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) InnerJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) NatualJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) CrossJoin(table interface{}, condition string) *Wrapper {
	return w
}

func (w *Wrapper) JoinWhere(table string, args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) JoinOrWhere(args ...interface{}) *Wrapper {
	return w
}

func (w *Wrapper) SubQuery(alias ...string) *Wrapper {
	return w
}

func (w *Wrapper) Has() (has bool, err error) {
	return
}

func (w *Wrapper) Disconnect() (err error) {
	return
}

func (w *Wrapper) Ping() (err error) {
	return
}

func (w *Wrapper) Connect() (err error) {
	return
}

func (w *Wrapper) Begin() *Wrapper {
	return w
}

func (w *Wrapper) Rollback() *Wrapper {
	return w
}

func (w *Wrapper) Commit() *Wrapper {
	return w
}

func (w *Wrapper) SetLockMethod() *Wrapper {
	return w
}

func (w *Wrapper) Lock() *Wrapper {
	return w
}

func (w *Wrapper) Unlock() *Wrapper {
	return w
}

func (w *Wrapper) SetQueryOption(options ...string) *Wrapper {
	return w
}

// Migration returns a new table migration struct
// based on the current database connection for the migration functions.
func (w *Wrapper) Migration() *Migration {
	return newMigration(w.db)
}
