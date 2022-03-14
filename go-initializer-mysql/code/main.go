package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aliyun/fc-runtime-go-sdk/fc"
	_ "github.com/go-sql-driver/mysql"
)

var MysqlDb *sql.DB
var MysqlDbErr error

func main() {
	fc.RegisterInitializerFunction(Init)
	fc.Start(HandleRequest)
}

// Init is initialize function, see https://help.aliyun.com/document_detail/323541.html for detail
func Init(ctx context.Context) {
	var (
		userName string = os.Getenv("DB_USER_NAME")
		password string = os.Getenv("DB_PASSWORD")
		endpoint string = os.Getenv("DB_ENDPOINT")
		port     string = os.Getenv("DB_PORT")
		database string = os.Getenv("DB_NAME")
		charset  string = "utf8"
	)

	if userName == "" || password == "" || endpoint == "" || database == "" || port == "" {
		MysqlDbErr = fmt.Errorf("database config is empty, %s, %s, %s, %s", userName, password, endpoint, database)
		return
	}

	// See https://github.com/go-sql-driver/mysql#dsn-data-source-name for how the DSN string is formatted
	mysqlDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s", userName, password, endpoint, port, database, charset)
	MysqlDb, MysqlDbErr = sql.Open("mysql", mysqlDsn)
	if MysqlDbErr != nil {
		log.Println("Data source name: " + mysqlDsn)
		return
	}

	// See "Important settings" section in https://github.com/go-sql-driver/mysql#important-settings
	MysqlDb.SetMaxOpenConns(10)
	MysqlDb.SetMaxIdleConns(10)
	MysqlDb.SetConnMaxLifetime(time.Minute * 3)

	pingCtx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if MysqlDbErr = MysqlDb.PingContext(pingCtx); nil != MysqlDbErr {
		log.Println("Data source name: " + mysqlDsn)
		log.Println("Connection to mysql fail: " + MysqlDbErr.Error())
	}
}

func HandleRequest() (*User, error) {
	return QueryOneDemo(10)
}


type User struct {
	ID   int64          `db:"id"`
	Name sql.NullString `db:"name"`
	Age  int            `db:"age"`
}

// QueryOneDemo: Just for example purpose.
func QueryOneDemo(limit int) (*User, error) {
	if MysqlDb == nil || MysqlDbErr != nil {
		panic(fmt.Errorf("MysqlDb initialize fail: %v", MysqlDbErr))
	}

	user := new(User)
	rows, err := MysqlDb.Query("select * from users limit ?", limit)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	if err != nil {
		log.Printf("Query failed, err: %v", err)
		return user, err
	}
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Name, &user.Age)
		if err != nil {
			log.Printf("Scan failed,err: %v", err)
			return user, err
		}
		log.Println(*user)
	}
	return user, nil
}
