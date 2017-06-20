# Reiner

一個由 Golang 撰寫且比起部分 ORM 還要讚的 MySQL 指令包覆函式庫。彈性高、不需要建構體標籤。實際上，這就只是 [PHP-MySQLi-Database-Class](https://github.com/joshcam/PHP-MySQLi-Database-Class) 不過是用在 Golang 而已（但還是多了些功能）。

#  這是什麼？

萊納是一個由 Golang 撰寫的 MySQL 的指令包覆函式庫（不是 ORM，永遠也不會是），幾乎所有東西都能操控於你手中。類似自己撰寫資料庫指令但是更簡單，JOIN 表格也變得比以前更方便了。

* 幾乎全功能的函式庫。
* 支援 MySQL 複寫橫向擴展機制（區分讀／寫連線）。
* 容易理解與記住、且使用方式十分簡單。
* SQL 指令建構函式。
* 資料庫表格建構協助函式。
* 支援子指令（Sub Query）。
* 可手動操作的交易機制（Transaction）和回溯（Rollback）功能。
* 透過預置聲明（Prepared Statement），99.9% 避免 SQL 注入攻擊。
* 自動脫逸表格名稱，避免觸動保留字。

# 為什麼？

[Gorm](https://github.com/jinzhu/gorm) 已經是 Golang 裡的 ORM 典範，但實際上要操作複雜與關聯性高的 SQL 指令時並不是很合適，而 Reiner 解決了這個問題。Reiner 也試圖不要和建構體扯上關係，不希望使用者需要手動指定任何標籤在建構體中。

# 執行緒與併發安全性？

我們都知道 Golang 的目標就是併發程式，當共用同個資料庫的時候請透過 `Copy()` 函式複製一份新的包覆函式庫，這能避免函式遭受干擾或覆寫。此方式並不會使資料庫連線遞增而造成效能問題，因此你可以有好幾個併發程式且有好幾個包覆函式庫的複製體都不會出現效能問題。

# 相關連結

這裡是 Reiner 受啟發，或是和資料庫有所關聯的連結。

* [kisielk/sqlstruct](http://godoc.org/github.com/kisielk/sqlstruct)
* [jmoiron/sqlx](https://github.com/jmoiron/sqlx)
* [russross/meddler](https://github.com/russross/meddler)
* [jinzhu/gorm](https://github.com/jinzhu/gorm)

# 安裝方式

打開終端機並且透過 `go get` 安裝此套件即可。

```bash
$ go get github.com/TeaMeow/Reiner
```

# 使用方式

Reiner 的使用方式十分直覺與簡易，類似基本的 SQL 指令集但是更加地簡化了。

## 資料庫連線

首先你需要透過函式來將 Reiner 連上資料庫，如此一來才能夠初始化包覆函式庫與相關的資料庫表格建構函式。一個最基本的單資料庫連線，讀寫都將透過此連線，連線字串共用於其它套件是基於 DSN（Data Source Name）。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/test?charset=utf8")
if err != nil {
    panic(err)
}
```

### 水平擴展（讀／寫分離）

這種方式可以有好幾個主要資料庫、副從資料庫，這意味著寫入時都會流向到主要資料庫，而讀取時都會向副從資料庫請求。這很適合用在大型結構還有水平擴展上。當你有多個資料庫來源時，Reiner 會逐一遞詢每個資料庫來源，英文稱其為 Round Robin，也就是每個資料庫都會輪流呼叫而避免單個資料庫負荷過重，也不會有隨機呼叫的事情發生。

```go
import "github.com/TeaMeow/Reiner"

db, err := reiner.New("root:root@/master?charset=utf8", []string{
	"root:root@/slave?charset=utf8",
	"root:root@/slave2?charset=utf8",
	"root:root@/slave3?charset=utf8",
})
if err != nil {
    panic(err)
}
```

## 資料綁定與處理

Reiner 允許你將結果與結構體切片或結構體綁定在一起。

```go
var user []*User
err := db.Bind(&user).Get("users")
```

### 逐行掃描

如果你偏好傳統的 `Row.Next` 來對每筆資料進行逐行掃描，Reiner 亦提供了 `Scan` 方式允許你傳入自訂的資料處理函式。你能夠在網路上找到ㄧ些輔助 `*sql.Rows` 的函式。

```go
err := db.Scan(func(rows *sql.Rows) {
	var username, password string
	rows.Scan(&username, &password)

	fmt.Println(username, password)
}).Get("users")
```

## 插入

透過 Reiner 你可以很輕鬆地透過建構體或是 `map` 來插入一筆資料。這是最傳統的插入方式，若該表格有自動遞增的編號欄位，插入後你就能透過 `LastInsertID` 獲得最後一次插入的編號。

```go
err := db.Insert("users", map[string]string{
	"username": "YamiOdymel",
	"password": "test",
})
// id := db.LastInsertID
```

### 覆蓋

```go
```

### 函式

插入時你可以透過 Reiner 提供的函式來執行像是 `SHA1()` 或者取得目前時間的 `NOW()`，甚至將目前時間加上一年⋯等。

```go
id, err := db.Insert("users", map[string]interface{}{
	"username":  "YamiOdymel",
	"password":  db.Func("SHA1(?)", "secretpassword+salt"),
	"expires":   db.Now("+1Y"),
	"createdAt": db.Now(),
})
```

### On Duplicate

```go
lastInsertID := "id"

id, err := db.OnDuplicate([]string{"updatedAt"}, lastInsertID).Insert("users", map[string]interface{}{
	"username":  "YamiOdymel",
	"password":  "test",
	"createdAt": db.Now(),
})
```

### 多筆資料

Reiner 允許你透過 `InsertMulti` 同時間插入多筆資料（單指令插入多筆資料），這省去了透過迴圈不斷執行單筆插入的困擾，這種方式亦大幅度提升了效能。

```go
data := []map[string]string{
	{
		"username": "YamiOdymel",
		"password": "test",
	}, {
		"username": "Karisu",
		"password": "12345",
	},
}

db.InsertMulti("users", data)
// ids := db.LastInsertIDs
```

## 更新

更新一筆資料在 Reiner 中極為簡單，你只需要指定表格名稱還有資料即可。

```go
db.Where("username", "YamiOdymel").Update("users", map[string]string{
	"username": "Karisu",
	"password": "123456",
})
// count := db.Count
```

### 筆數限制

`Limit` 能夠限制更新的筆數，如果是 `10`，那就表示只更新最前面 10 筆資料而非全部。

```go
db.Limit(10).Update("users", data)
```

## 選擇與取得

最基本的選擇在 Reiner 中稱之為 `Get` 而不是 `Select`。如果你想要取得 `rows.Next` 來掃描每一行的結果，Reiner 提供了 `LastRows` 即為最後一次的 `rows` 資料。

```go
// 等效於：SELECT * FROM users
err := db.Get("users")
// rows := db.LastRows
// for rows.Next() {
//     rows.Scan(...)
// }
```

### 筆數限制

`Limit` 能夠限制取得的筆數，如果是 `10`，那就表示只取得最前面 10 筆資料而非全部。

```go
// 等效於：SELECT * FROM users LIMIT 10
db.Limit(10).Get("users")
```

### 指定欄位

你可以透過 `Columns` 指定要取得的欄位名稱，多個欄位由逗點區分，亦能是函式。

```go
// 等效於：SELECT username, nickname FROM users
db.Columns("username", "nickname").Get("users")
// 等效於：SELECT COUNT(*) AS count FROM users
db.Columns("COUNT(*) AS count").Get("users")
```

### 單行資料

預設來說 `Get` 會回傳一個切片或是陣列，這令你需要透過迴圈逐一取得資料，但某些情況下你很確信你僅要取得一筆資料的話，可以嘗試 `GetOne`。這能將資料直接映射到單個建構體上而避免你需要透過迴圈處理資料的麻煩。

```go
db.Where("id", 1).GetOne("users")
// 或者像這樣使用函式。
db.Columns("SUM(id)", "COUNT(*) AS cnt").GetOne("users")
```

### 取得單值

這就像 `GetOne`，但 `GetValue` 取得的是單個欄位的內容，例如說你想要單個使用者的暱稱，甚至是多個使用者的暱稱陣列就很適用。

```go
db.Columns("username").GetValue("users")
// 也能搭配 Limit。
db.Limit(5).Columns("username").GetValue("users")
// 或者是函式。
db.Columns("COUNT(*)").GetValue("users")
```

### 分頁功能

分頁就像是取得資料ㄧ樣，但更擅長用於多筆資料、不會一次顯示完畢的內容。Reiner 能夠幫你自動處理換頁功能，讓你不需要自行計算換頁時的筆數應該從何開始。為此，你需要定義兩個變數，一個是目前的頁數，另一個是單頁能有幾筆資料。

```go
page := 1
db.PageLimit = 2

db.Paginate("users", page)
// fmt.Println("目前頁數為 %d，共有 %d 頁", page, db.TotalPages)
```

## 執行生指令

Reiner 已經提供了近乎日常中 80% 會用到的方式，但如果好死不死你想使用的功能在那 20% 之中，我們還提供了原生的方法能讓你直接輸入 SQL 指令執行自己想要的鳥東西。一個最基本的生指令（Raw Query）就像這樣。

其中亦能帶有預置聲明（Prepared Statement），也就是指令中的問號符號替代了原本的值。這能避免你的 SQL 指令遭受注入攻擊。

```go
db.RawQuery("SELECT * from users WHERE id >= ?", 10)
```

### 單行資料

僅選擇單筆資料的生指令函式，這意味著你能夠將取得的資料直接映射到一個建構體上。

```go
db.RawQueryOne("SELECT * FROM users WHERE id = ?", 10)
```

### 取得單值

透過 `RawQueryValue` 可以直接取得單個欄位得值，而不是一個陣列或切片。

```go
db.RawQueryValue("SELECT password FROM users WHERE id = ? LIMIT 1", 10)
```

### 單值多行

透過 `RawQueryValue` 能夠取得單一欄位的值，當有多筆結果的時候會取得一個值陣列。

```go
db.RawQueryValue("SELECT username FROM users LIMIT 10")
```

### 進階方式

如果你對 SQL 指令夠熟悉，你也可以使用更進階且複雜的用法。

```go
err := db.RawQuery("SELECT id, firstName, lastName FROM users WHERE id = ? AND username = ?", 1, "admin")

// will handle any SQL query.
params := []int{10, 1, 10, 11, 2, 10}
query := `(
    SELECT a FROM t1
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
) UNION (
    SELECT a FROM t2
        WHERE a = ? AND B = ?
        ORDER BY a LIMIT ?
)`
err := db.RawQuery(query, params...)
```

## 條件宣告

透過 Reiner 宣告 `WHERE` 條件也能夠很輕鬆。一個最基本的 `WHERE AND` 像這樣使用。

```go
db.Where("id", 1).Where("username", "admin").Get("users")
// 等效於：SELECT * FROM users WHERE id=1 AND username='admin';
```

### 擁有

`HAVING` 能夠與 `WHERE` 一同使用。

```go
db.Where("id", 1).Having("username", "admin").Get("users")
// 等效於：SELECT * FROM users WHERE id=1 HAVING username='admin';
```

### 欄位比較

如果你想要在條件中宣告某個欄位是否等於某個欄位⋯你能夠像這樣。

```go
// 別這樣。
db.Where("lastLogin", "createdAt").Get("users")
// 這樣才對。
db.Where("lastLogin = createdAt").Get("users")
// 等效於：SELECT * FROM users WHERE lastLogin = createdAt;
```

### 自訂運算子

在 `Where` 或 `Having` 的最後一個參數你可以自訂條件的運算子，如 `>=`、`<=`、`<>`⋯等。

```go
db.Where("id", 50, ">=").Get("users")
// 等效於：SELECT * FROM users WHERE id >= 50;
```

### 介於／不介於

透過 `BETWEEN` 和 `NOT BETWEEN` 條件也可以用來限制數值內容是否在某數之間（相反之，也能夠限制是否不在某範圍內）。

```go
db.Where("id", []int{0, 20}, "BETWEEN").Get("users")
// 等效於：SELECT * FROM users WHERE id BETWEEN 4 AND 20
```

### 於清單／不於清單內

透過 `IN` 和 `NOT IN` 條件能夠限制並確保取得的內容不在（或者在）指定清單內。

```go
db.Where("id", []interface{}{1, 5, 27, -1, "d"}, "IN").Get("users")
// 等效於：SELECT * FROM users WHERE id IN (1, 5, 27, -1, 'd');
```

### 或／還有或

通常來說多個 `Where` 會產生 `AND` 條件，這意味著所有條件都必須符合，有些時候你只希望符合部分條件即可，就能夠用上 `OrWhere`。

```go
db.Where("firstName", "John").OrWhere("firstName", "Peter").Get("users")
// 等效於：SELECT * FROM users WHERE firstName='John' OR firstName='peter'
```

如果你的要求比較多，希望達到「A = B 或者 (A = C 或 A = D)」的話，你可以嘗試這樣。

```go
db.Where("A = B").OrWhere("(A = C OR A = D)").Get("users")
// 等效於：SELECT * FROM users WHERE A = B OR (A = C OR A = D)
```

### 空值

要確定某個欄位是否為空值，傳入一個 `nil` 即可。

```go
// 別這樣。
db.Where("lastName", "NULL", "IS NOT").Get("users")
// 這樣才對。
db.Where("lastName", nil, "IS NOT").Get("users")
// 等效於：SELECT * FROM users where lastName IS NOT NULL
```

### Raw

```go
db.Where("id != companyId").Where("DATE(createdAt) = DATE(lastLogin)").Get("users")
// 等效於：SELECT * FROM users WHERE id != companyId AND DATE(createdAt) = DATE(lastLogin)
```

### Raw With Params

```go
db.Where("(id = ? or id = ?)", []int{6, 2}).Where("login", "mike").Get("users")
// 等效於：SELECT * FROM users WHERE (id = 6 or id = 2) and login='mike';
```

## 刪除

刪除一筆資料再簡單不過了，透過 `Count` 計數能夠清楚知道你的 SQL 指令影響了幾行資料，如果是零的話即是無刪除任何資料。

```go
err := db.Where("id", 1).Delete("users")
if err == nil && db.Count != 0 {
    fmt.Println("成功地刪除了一筆資料！")
}
```

## 排序

Reiner 亦支援排序功能，如遞增或遞減，亦能擺放函式。

```go
db.OrderBy("id", "ASC").OrderBy("login", "DESC").OrderBy("RAND()").Get("users")
// 等效於：SELECT * FROM users ORDER BY id ASC, login DESC, RAND();
```

### 從值排序

也能夠從值進行排序，只需要傳入一個切片即可。

```go
db.OrderBy("userGroup", "ASC", []string{"superuser", "admin", "users"}).Get("users")
// 等效於：SELECT * FROM users ORDER BY FIELD (userGroup, 'superuser', 'admin', 'users') ASC;
```

## 群組

簡單的透過 `GroupBy` 就能夠將資料由指定欄位群組排序。

```go
db.GroupBy("name").Get("users")
// 等效於：SELECT * FROM users GROUP BY name;
```

## 加入

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.Where("u.id", 6)
db.Get("products p", "u.name, p.productName")
```

### 條件限制

```go
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinWhere("users u", "u.tenantID", 5)
db.Get("products p", "u.name, p.productName")
// 等效於：SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID AND u.tenantID = 5)
```

```go
// db.InnerJoin()
// db.LeftJoin()
// db.RightJoin()
// db.NaturalJoin()
// db.CrossJoin()
db.Join("users u", "p.tenantID = u.tenantID", "LEFT")
db.JoinOrWhere("users u", "u.tenantID", "=", 5)

err := db.Get("products p", "u.name, p.productName")
// 等效於：SELECT u.login, p.productName FROM products p LEFT JOIN users u ON (p.tenantID=u.tenantID OR u.tenantID = 5)
```

## 子指令

```go
subQuery := db.SubQuery()
subQuery.Get("users")
```

```go
subQuery := db.SubQuery("sq")
subQuery.Get("users")
```

### 選擇／取得

```go
idSubQuery := db.SubQuery()
idSubQuery.Where("qty", 2, ">").Get("products", "userId")

err := db.Where("id", idSubQuery, "IN").Get("users")
// 等效於：SELECT * FROM users WHERE id IN (SELECT userId FROM products WHERE qty > 2)
```

### 插入

```go
idSubQuery := db.SubQuery()
idSubQuery.Where("id", 6).GetOne("users", "name")

err := db.Insert("products", map[string]interface{}{
	"productName": "test product",
	"userID":      idSubQuery,
	"lastUpdated": db.Now(),
})
// 等效於：INSERT INTO PRODUCTS (productName, userId, lastUpdated) values ("test product", (SELECT name FROM users WHERE id = 6), NOW());
```

### 加入

```go
userSubQuery := db.SubQuery("u")
userSubQuery.Where("active", 1).Get("users")

err := db.Join(userSubQuery, "p.userId = u.id", "LEFT").Get("products p", "u.login, p.productName")
// 等效於：SELECT u.login, p.productName FROM products p LEFT JOIN (SELECT * FROM t_users WHERE active = 1) u on p.userId=u.id;
```

### 存在／不存在

```go
subQuery := db.SubQuery()
subQuery.Where("company", "testCompany")
subQuery.Get("users", "userId")

err := db.Where("", subQuery, "EXISTS").Get("products")
// 等效於：SELECT * FROM products WHERE EXISTS (select userId from users where company='testCompany')
```

## 是否擁有該筆資料

有些時候我們只想知道資料庫是否有符合的資料，但並不是要取得其資料，舉例來說就像是登入是僅是要確認帳號密碼是否吻合，此時就可以透過 `Has` 用來確定資料庫是否有這筆資料。

```go
db.Where("username", "yamiodymel").Where("password", "123456")
has, err := db.Has("users")
if has {
	fmt.Println("登入成功！")
} else {
	fmt.Println("帳號或密碼錯誤。")
}
```

## 輔助函式

Reiner 有提供一些輔助用的函式協助你除錯、紀錄，或者更加地得心應手。

### 資料庫連線

透過 `Disconnect` 結束一段連線。

```go
db.Disconnect()
```

你也能在資料庫發生錯誤、連線遺失時透過 `Connect` 來重新手動連線。

```go
if err := db.Ping(); err != nil {
	db.Connect()
}
```

### 最後執行的 SQL 指令

取得最後一次所執行的 SQL 指令，這能夠用來記錄你所執行的所有動作。

```go
db.Get("users")
fmt.Println("最後一次執行的 SQL 指令是：%s", db.LastQuery)
```

### 結果／影響的行數

行數很常用於檢查是否有資料、作出變更。資料庫不會因為沒有變更任何資料而回傳一個錯誤（資料庫僅會在真正發生錯誤時回傳錯誤資料），所以這是很好的檢查方法。

```go
db.Get("users")
fmt.Println("總共獲取 %s 筆資料", db.Count)
db.Delete("users")
fmt.Println("總共刪除 %s 筆資料", db.Count)
db.Update("users", data)
fmt.Println("總共更新 %s 筆資料", db.Count)
```

## 交易函式

交易函式僅限於 InnoDB 型態的資料表格，這能令你的資料寫入更加安全。你可以透過 `Begin` 開始記錄並繼續你的資料庫寫入行為，如果途中發生錯誤，你能透過 `Rollback` 回到紀錄之前的狀態，即為回溯（或滾回、退回），如果這筆交易已經沒有問題了，透過 `Commit` 將這次的變更永久地儲存到資料庫中。

```go
err := db.Begin().Insert("myTable", data)
if err != nil {
	db.Rollback()
} else {
	db.Commit()
}
```

## 鎖定表格

你能夠手動鎖定資料表格，避免同時間寫入相同資料而發生錯誤。

```go
db.SetLockMethod("WRITE").Lock("users")

// 呼叫其他的 `Lock()` 函式也會自動將前一個上鎖解鎖，當然你也可以手動呼叫 `Unlock()` 解鎖。
db.Unlock()

// 同時間要鎖上兩個表格也很簡單。
db.SetLockMethod("READ").Lock("users", "log")
```

## 指令關鍵字

Reiner 也支援設置指令關鍵字。

```go
db.SetQueryOption("LOW_PRIORITY").Insert("users", data)
// 等效於：INSERT LOW_PRIORITY INTO table ...

db.SetQueryOption("FOR UPDATE").Get("users")
// 等效於：SELECT * FROM users FOR UPDATE;

db.SetQueryOption("SQL_NO_CACHE").Get("users")
// 等效於：GIVES: SELECT SQL_NO_CACHE * FROM users;
```

### 多個選項

你亦能同時設置多個關鍵字給同個指令。

```go
db.SetQueryOption("LOW_PRIORITY", "IGNORE").Insert("users", data)
// Gives: INSERT LOW_PRIORITY IGNORE INTO users ...
```

# 表格建構函式

Reiner 除了基本的資料庫函式可供使用外，還能夠建立一個表格並且規劃其索引、外鍵、型態。

```go
migration := db.Migration()

migration.Column("test").Varchar(32).Primary().CreateTable("test_table")
// 等效於：CREATE TABLE `test_table` (`test` varchar(32) NOT NULL PRIMARY KEY) ENGINE=INNODB
```


| 數值       | 字串       | 二進制     | 檔案資料     | 時間      | 浮點數     | 固組   |
|-----------|------------|-----------|------------|-----------|-----------|-------|
| TinyInt   | Char       | Binary    | Blob       | Date      | Double    | Enum  |
| SmallInt  | Varchar    | VarBinary | MediumBlob | DateTime  | Decimal   | Set   |
| MediumInt | TinyText   | Bit       | LongBlob   | Time      | Float     |       |
| Int       | Text       |           |            | Timestamp |           |       |
| BigInt    | MediumText |           |            | Year      |           |       |
|           | LongText   |           |            |           |           |       |