package backup

import (
//	"../../interface/databases"
//	"log"
)

type BackupMap struct {
	offset  int
	size    int
	getdata func() ([]byte, error)
	putdata func([]byte)
}

type BackupCronT []BackupMap

var BackupCron = make(BackupCronT, 0)

func (self BackupCronT) AddTask(getter func() ([]byte, error), putter func([]byte)) {
}

func (self BackupCronT) Restore(dumpfile string) error {
	return nil
}

func PackData(Data interface{}) {
	/*	dbs := databases.GetDbListForBackup()
		buf, err := dbs.Encode()
		if err != nil {
			log.Println(err.Error())
		}
		log.Println(buf)
		log.Println(string(buf))*/
}
