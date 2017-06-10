# Reiner

A Golang database ORM with the 1990's style. Flexible, and no struct tags needed. More actually, it's just [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) but in Golang.

# What is it?

A MySQL ORM written in Golang which lets you controll everything, just like writing a query but simpler, join tables are now easier than before.

* Almost full-featured ORM
* Easy to remember, understand
* SQL Builder
* Table Migrations
* Sub queries
* Transactions

# Why?

[Gorm](https://github.com/jinzhu/gorm) is great as fuck, but it's not really fits with a complex SQL query usage, and Reiner solved the problem. Reiner also decoupling the function usage with the struct (Loose coupling).

# Thread Safe?

# Installtion

```bash
$ go get github.com/TeaMeow/Reiner
```

# Helper Types

```go
// F for fields.
type F map[string]interface{}

// Fs for field group.
type Fs []F{}

// V for values.
type V []interface{}

// O for options.
type O struct {

}
```

# Usage

## Conenction

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("yamiodymel:yamiodymel@/test?charset=utf8")
if err != nil {
    panic(err)
}
```



## Insert

### Traditional/Replace

```go
err := db.Insert("users", reiner.Fields{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

### Functions

```go
err := db.Insert("users", reiner.Fields{
	"username":  "YamiOdymel",
	"password":  db.Func("SHA1(?)", reiner.Values{"secretpassword+salt"}),
	"expires":   db.Now("+1Y"),
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

### On Duplicate

```go
lastInsertID := "id"

err := db.Columns("updatedAt").OnDuplicate(lastInsertID).Insert("users", reiner.Fields{
	"username":  "YamiOdymel",
	"password":  "test",
	"createdAt": db.Now(),
})
// id := db.LastInsertID
```

### Multiple

```go
data := reiner.FieldGroup{
	reiner.Fields{
		"username": "YamiOdymel",
		"password": "test",
	},
	reiner.Fields{
		"username": "Karisu",
		"password": "12345",
	},
}

err := db.InsertMulti("users", data)
// ids := db.LastInsertIDs
```



## Update

```go
err := db.Where("username", "YamiOdymel").Update("users", reiner.Fields{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

### Limit

```go
err := db.Limit(10).Update("users", data)
```



## Select

```go
err := db.Bind(&users).Get("users")
```

### Limit

```go
err := db.Bind(&users).Limit(10).Get("users")
```

### Specified Columns

```go
err := db.Bind(&users).Columns("username", "nickname").Get("users")
// count := db.Count
```

### Single Row

```go
err := db.Bind(&user).Where("id", 1).GetOne("users")
// or with the custom query.
err := db.Bind(&stats).GetOne("users", reiner.Option{
	Query: "sum(id), count(*) as cnt",
})
```

### Get Value

```go
err := db.Bind(&usernames).GetValue("users", "username")
// or with the limit.
err := db.Bind(&usernames).Limit(5).GetValue("users", "username")
// or with the function.
err := db.Bind(&total).GetValue("users", "count(*)")
```

### Paginate

```go
page := 1
db.PageLimit = 2

err := db.Bind(&users).Paginate("users", page)
// fmt.Println("Showing %d out of %d", page, db.TotalPages)
```



## Raw Queries

### Common

```go
err := db.Bind(&users).RawQuery("SELECT * from users WHERE id >= ?", reiner.Values{10})
```

### Single Row

```go
err := db.Bind(&user).RawQueryOne("SELECT * FROM users WHERE id = ?", reiner.Values{10})
```

### Single Value

```go
err := db.Bind(&password).RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", reiner.Values{10})
```

### Single Value From Multiple Rows

```go
err := db.Bind(&usernames).RawQueryValue("SELECT username FROM users LIMIT 10")
```

### Advanced

```go
params := reiner.Values{1, "admin"}
err := db.Bind(&users).RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", params)

// will handle any SQL query.
params = reiner.Values{10, 1, 10, 11, 2, 10}
query := "(
    SELECT a FROM t1
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
) UNION (
    SELECT a FROM t2
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
)"
err := db.Bind(&results).RawQuery(query, params)
```



## Conditions

### Equals

```go
db.Where("id", 1)
db.Where("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 AND username='admin';
```

#### Having

```go
db.Where("id", 1)
db.Having("username", "admin")
db.Bind(&users).Get("users")

// Equals: SELECT * FROM users WHERE id=1 HAVING username='admin';
```

#### Columns Comparison

```go
// WRONG
db.Where("lastLogin", "createdAt")
// CORRECT
db.Where("lastLogin = createdAt")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE lastLogin = createdAt;
```

### Custom

```go
db.Bind(&users).Where("id", 50, ">=").Get("users")
// Equals: SELECT * FROM users WHERE id >= 50;
```

### Between / Not Between

```go
db.Bind(&users).Where("id", reiner.Values{0, 20}, "BETWEEN").Get("users")
// Equals: SELECT * FROM users WHERE id BETWEEN 4 AND 20
```

### In / Not In

```go
db.Bind(&users).Where("id", reiner.Values{1, 5, 27, -1, "d"}, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (1, 5, 27, -1, 'd');
```

### Or / And Or

```go
db.Where("firstName", "John")
db.OrWhere("firstName", "Peter")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE firstName='John' OR firstName='peter'
```

### Null

```go
db.Where("lastName", reiner.NULL, "IS NOT")
db.Bind(&users).Get("users")
// Equals: SELECT * FROM users where lastName IS NOT NULL
```

### Raw

```go
db.Where("id != companyId")
db.Where("DATE(createdAt) = DATE(lastLogin)")
db.Bind(&users).Get("users")
```

### Raw With Params

```go
db.Where("(id = ? or id = ?)", reiner.Values{6, 2})
db.Where("login", "mike")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users WHERE (id = 6 or id = 2) and login='mike';
```



## Delete

### Common

```go
err := db.Where("id", 1).Delete("users")
if err == nil && db.Count != 0 {
    fmt.Println("Deleted successfully!")
}
```



## Order

```go
db.OrderBy("id", "ASC")
db.OrderBy("login", "DESC")
db.OrderBy("RAND ()")

db.Bind(&users).Get("users")
// Equals: SELECT * FROM users ORDER BY id ASC,login DESC, RAND ();
```

### By Values

```go
db.OrderBy("userGroup", "ASC", []string{"superuser", "admin", "users"})
db.Bind(&users).Get("users")
// Equals: SELECT * FROM users ORDER BY FIELD (userGroup, 'superuser', 'admin', 'users') ASC;
```



## Group

```go
db.GroupBy("name").Bind(&users).Get("users")
// Equals: SELECT * FROM users GROUP BY name;
```



## Join

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.Where("u.id", 6)

db.Bind(&products).Get("products p", "u.name, p.productName")
```

### Conditions

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinWhere("users u", "u.tenantID", 5)

db.Bind(&products).Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID AND u.tenantID = 5)
```

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinOrWhere("users u", "u.tenantID", 5)

db.Bind(&products).Get("products p", "u.name, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID OR u.tenantID = 5)
```



## Subqueries

```go
sq := db.SubQuery()
sq.Get("users")
```

```go
sq := db.SubQuery("sq")
sq.Get("users")
```

### Select

```go
ids := db.SubQuery()
ids.Where("qty", 2, ">").Get("products", "userId")

db.Where("id", ids, "IN").Get("users")
// Equals: SELECT * FROM users WHERE id IN (SELECT userId FROM products WHERE qty > 2)
```

### Insert

```go
userIDQ := db.subQuery()
userIDQ.Where("id", 6).GetOne("users", "name")

err := db.insert("products", reiner.Fields{
	"productName": "test product",
	"userID":      userIDQ,
	"lastUpdated": db.Now(),
})
// Equals: INSERT INTO PRODUCTS (productName, userId, lastUpdated) values ("test product", (SELECT name FROM users WHERE id = 6), NOW());
```

### Join

```go
usersQ := db.SubQuery("u")
userQ.Where("active", 1).Get("users")

db.Join(usersQ, "p.userId = u.id", "LEFT").Get("products p", "u.login, p.productName")
// Equals: SELECT u.login, p.productName FROM products p LEFT JOIN (SELECT * FROM t_users WHERE active = 1) u on p.userId=u.id;
```

### Exist / Not Exist

```go
sub := db.SubQuery()
sub.Where("company", "testCompany")
sub.Get("users", "userId")

db.Where("", sub, "EXISTS").Get("products")
// Equals: SELECT * FROM products WHERE EXISTS (select userId from users where company='testCompany')
```



## Has

```go
db.Where("username", "yamiodymel").Where("password", "123456")

if db.Has("users") {
	fmt.Println("Logged in successfully!")
} else {
	fmt.Println("Incorrect username or the password.")
}
```



## Helpers

```go
db.Disconnect()
```

```go
if !db.Ping() {
	db.Connect()
}
```

```go
db.Get("users")

fmt.Println("Last executed query was %s", db.LastQuery())
```



## Transactions

```go
err := db.Begin().Insert("myTable", data)
if err != nil {
	db.Rollback()
} else {
	db.Commit()
}
```



## Lock

```go
db.SetLockMethod("WRITE").Lock("users")

// Calling another `Lock()` will unlock the first lock. You could also use `Unlock()`
db.Unlock()

// Lock the multiple tables at the same time is easy.
db.SetLockMethod("READ")->Lock("users", "log")
```



## Query Keywords

### Common

```go
db.SetQueryOption("LOW_PRIORITY").Insert("users", data)
// Equals: INSERT LOW_PRIORITY INTO table ...

db.SetQueryOption("FOR UPDATE").Get("users")
// Equals: SELECT * FROM users FOR UPDATE;

db.SetQueryOption("SQL_NO_CACHE").Get("users")
// Equals: GIVES: SELECT SQL_NO_CACHE * FROM users;
```

### Multiple

```go
db.SetQueryOption("LOW_PRIORITY", "IGNORE").Insert("users", data)
// GIVES: INSERT LOW_PRIORITY IGNORE INTO users ...
```

# Table Migrations

```go
migration := db.Migration()
migration.Column("test").Varchar(32).Primary().CreateTable("test_table")
```