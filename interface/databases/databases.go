package databases

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
)

type Db struct {
	Ip       string
	User     string
	Password string
	Port     string
	Group    string
}

type DbDescr struct {
	Ip    string
	User  string
	Port  string
	Group string
}

type DbMap map[string]*Db
type DbLst map[string]DbMap
type HostMap map[string]*Db

var databaseList = make(DbLst)
var HostsList = make(map[string]*Db)

func (self Db) GetDescription() DbDescr {
	return DbDescr{Ip: self.Ip, User: self.User, Port: self.Port, Group: self.Group}
}

func (self DbLst) Encode() ([]byte, error) {
	gobBuffer := new(bytes.Buffer)
	binBuffer := new(bytes.Buffer)
	gobEnc := gob.NewEncoder(gobBuffer)
	if err := gobEnc.Encode(self); err != nil {
		return nil, err
	}
	if err := binary.Write(binBuffer, binary.BigEndian, gobBuffer.Bytes()); err != nil {
		return nil, err
	}
	return binBuffer.Bytes(), nil
}

func AddReplicaGroup(name string) error {
	if _, ok := databaseList[name]; ok {
		return errors.New("Already exists")
	}
	databaseList[name] = make(DbMap)
	return nil
}

func EditReplicaGroup(oldname string, newname string) error {
	if _, ok := databaseList[newname]; ok {
		return errors.New("Already exists")
	}
	if _, ok := databaseList[oldname]; !ok {
		return errors.New(oldname + " replication group wasn't found")
	}
	databaseList[newname] = databaseList[oldname]
	DeleteReplicaGroup(oldname)
	return nil
}

func DeleteReplicaGroup(name string) {
	for key, _ := range databaseList[name] {
		delete(HostsList, key)
	}
	delete(databaseList, name)
}

func AddMySQLHost(db Db) (DbDescr, error) {
	if _, ok := databaseList[db.Group]; !ok {
		return DbDescr{}, errors.New("Replication Group wasn't found")
	}
	if _, ok := databaseList[db.Group][db.Ip+":"+db.Port]; ok {
		return DbDescr{}, errors.New("Host already exists")
	}
	if _, ok := HostsList[db.Ip+":"+db.Port]; ok {
		return DbDescr{}, errors.New("Host already exists")
	}
	NewDb := Db{User: db.User, Password: db.Password, Port: db.Port, Ip: db.Ip, Group: db.Group}
	databaseList[db.Group][db.Ip+":"+db.Port] = &NewDb
	HostsList[db.Ip+":"+db.Port] = databaseList[db.Group][db.Ip+":"+db.Port]
	return NewDb.GetDescription(), nil
}

func DeleteMySQLHost(name string) error {
	if _, ok := HostsList[name]; !ok {
		return errors.New("Host " + name + " wasn't found")
	}
	group := HostsList[name].Group
	delete(databaseList[group], name)
	delete(HostsList, name)
	return nil
}

func GetDatabasesList() map[string][]Db {
	var res = make(map[string][]Db)
	for key, _ := range databaseList {
		res[key] = make([]Db, 0)
		for _, val := range databaseList[key] {
			res[key] = append(res[key], *val)
		}
	}
	return res
}

func GetDbListForBackup() ([]byte, error) {
	res, err := databaseList.Encode()
	if err != nil {
		return nil, err
	}
	return res, nil
}
