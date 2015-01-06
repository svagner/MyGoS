package mysql

import (
	"../../interface/databases"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type masterInfo struct {
	File             string
	Position         string
	Binlog_Do_DB     string
	Binlog_Ignore_DB string
}

type slaveInfo struct {
}

type DatabasesStat struct {
	Ip          string
	Group       string
	SlaveStatus string
	LastError   string
	Uptime      string
	Master      masterInfo
	Slave       slaveInfo
}

type StatMap map[string]DatabasesStat

func GetMySQLInfo() (StatMap, error) {
	var databasesStat = make(StatMap)
	for _, dbHost := range databases.HostsList {
		statRes := new(DatabasesStat)
		db, err := sql.Open("mysql", dbHost.User+":"+dbHost.Password+"@tcp("+dbHost.Ip+":"+dbHost.Port+")/drugvokrug?timeout=1s")
		defer db.Close()
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		var got string // <- for non-use data
		// get slave info
		err = db.QueryRow("show slave status;").Scan(&got)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		// get uptime
		err = db.QueryRow("show status like 'Uptime';").Scan(&got, &statRes.Uptime)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		// Get master info
		err = db.QueryRow("show master status;").Scan(&statRes.Master.File, &statRes.Master.Position, &statRes.Master.Binlog_Do_DB, &statRes.Master.Binlog_Ignore_DB)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		databasesStat[dbHost.Ip+":"+dbHost.Port] = *statRes
	}
	return databasesStat, nil
}
