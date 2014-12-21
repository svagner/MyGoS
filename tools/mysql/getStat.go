package mysql

import (
	"../../interface/databases"
	//	"../../interface/events"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type DatabasesStat struct {
	Ip          string
	Group       string
	SlaveStatus string
}

func GetMySQLInfo(host string, port string) error {
	log.Println(host)
	if _, ok := databases.HostsList[host+":"+port]; !ok {
		return errors.New("Host " + host + ":" + port + " wasn't found in internal data records")
	}
	db, err := sql.Open("mysql", databases.HostsList[host+":"+port].User+":"+databases.HostsList[host+":"+port].Password+"@tcp("+host+":"+port+")/drugvokrug")
	defer db.Close()
	if err != nil {
		return err
	}
	rows, err := db.Query("show slave status;")
	if err != nil {
		return err
	}
	for rows.Next() {
		var got string
		rows.Scan(&got)
		log.Println(got)
	}
	return nil
}
