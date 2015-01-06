package databases

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"os/exec"
	"sort"
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

const (
	STEP_INTERNAL = 0x1
	STEP_SCRIPT   = iota
)

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
	gobEnc := gob.NewEncoder(gobBuffer)
	if err := gobEnc.Encode(self); err != nil {
		return nil, err
	}
	return gobBuffer.Bytes(), nil
}

func (self *DbLst) Decode(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(self); err != nil {
		return err
	}
	for key, _ := range *self {
		for host, _ := range (*self)[key] {
			HostsList[host] = databaseList[key][host]
		}
	}
	return nil
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

func RestoreDbListFromBackup(data []byte) error {
	err := databaseList.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

// Replication steps
type MySQLReplicaStep struct {
	Name     string
	Type     byte
	Content  string
	Changer  func(string) ([]byte, error)
	Rollback func()
	Pos      int
}

/* Replica step methods */

func (self MySQLReplicaStep) Run() (result []byte, err error) {
	result, err = self.Changer(self.Content)
	return
}

/* end */

type MySQLReplicaStepArray []MySQLReplicaStep

/* Sort methods for type MySQLReplicaStepArray */
func (self MySQLReplicaStepArray) Len() int {
	return len(self)
}

func (self MySQLReplicaStepArray) Less(i, j int) bool {
	return self[i].Pos < self[j].Pos
}

func (self MySQLReplicaStepArray) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

/* end */

type MySQLReplicaScript struct {
	Name    string
	Content string
}

type _mySQLReplStepsHash map[string]*MySQLReplicaStep

/* Type's methods */

func (self _mySQLReplStepsHash) ToList() MySQLReplicaStepArray {
	res := make([]MySQLReplicaStep, 0)
	for _, value := range self {
		res = append(res, *value)
	}
	return res
}

func (self *_mySQLReplStepsHash) FromList(data MySQLReplicaStepArray) {
	for _, value := range data {
		if _, ok := (*self)[value.Name]; !ok {
			(*self)[value.Name] = &value
		} else {
			(*self)[value.Name].Pos = value.Pos
		}
		if (*self)[value.Name].Type == STEP_SCRIPT {
			(*self)[value.Name].Changer = runScript
		}
	}
}

func (self _mySQLReplStepsHash) GetSelected() (count uint, selected MySQLReplicaStepArray) {
	for _, value := range self {
		if value.Pos >= 0 {
			selected = append(selected, *value)
			count++
		}
	}
	return
}

/* end */

/* Replication steps functions */
func runScript(content string) (res []byte, err error) {
	cmd := exec.Command("/bin/sh", "-c", content)
	res, err = cmd.Output()
	return
}

/* end */

var mySQLReplStepsHash = _mySQLReplStepsHash{
	"Enable RO Mode at master node":       &MySQLReplicaStep{"Enable RO Mode at master node", STEP_INTERNAL, "", nil, nil, -1},
	"Check slave thread status at slave":  &MySQLReplicaStep{"Check slave thread status at slave", STEP_INTERNAL, "", nil, nil, -1},
	"Check slave thread status at master": &MySQLReplicaStep{"Check slave thread status at master", STEP_INTERNAL, "", nil, nil, -1},
	"Stop slave thread at slave node":     &MySQLReplicaStep{"Stop slave thread at slave node", STEP_INTERNAL, "", nil, nil, -1},
	"Start slave thread at master node":   &MySQLReplicaStep{"Start slave thread at master node", STEP_INTERNAL, "", nil, nil, -1},
	"Disable RO Mode at slave node":       &MySQLReplicaStep{"Disable RO Mode at slave node", STEP_INTERNAL, "", nil, nil, -1},
	"Stop slave thread at master node":    &MySQLReplicaStep{"Stop slave thread at master node", STEP_INTERNAL, "", nil, nil, -1},
}

var myReplChoosenSteps MySQLReplicaStepArray

func GetChoosenReplicaSteps() MySQLReplicaStepArray {
	if len(myReplChoosenSteps) == 0 {
		myReplChoosenSteps = append(myReplChoosenSteps, *mySQLReplStepsHash["Enable RO Mode at master node"])
		mySQLReplStepsHash["Enable RO Mode at master node"].Pos = 0
		myReplChoosenSteps = append(myReplChoosenSteps, *mySQLReplStepsHash["Check slave thread status at slave"])
		mySQLReplStepsHash["Check slave thread status at slave"].Pos = 1
		myReplChoosenSteps = append(myReplChoosenSteps, *mySQLReplStepsHash["Check slave thread status at master"])
		mySQLReplStepsHash["Check slave thread status at master"].Pos = 2
	}
	return myReplChoosenSteps
}

func GetReplicaStepsForChoose() MySQLReplicaStepArray {
	res := make([]MySQLReplicaStep, 0)
	for _, step := range mySQLReplStepsHash {
		res = append(res, *step)
	}
	return res
}

func AddReplicationStep(data MySQLReplicaScript) {
	mySQLReplStepsHash[data.Name] = &MySQLReplicaStep{Name: data.Name, Content: data.Content, Type: STEP_SCRIPT, Changer: runScript, Rollback: nil, Pos: -1}
}

func DeleteReplicationStep(name string) {
	if mySQLReplStepsHash[name].Pos != -1 {
		myReplChoosenSteps = concatReplSteps(myReplChoosenSteps[:mySQLReplStepsHash[name].Pos], myReplChoosenSteps[(mySQLReplStepsHash[name].Pos+1):])
	}
	delete(mySQLReplStepsHash, name)
}

func concatReplSteps(old1, old2 []MySQLReplicaStep) MySQLReplicaStepArray {
	newslice := make(MySQLReplicaStepArray, len(old1)+len(old2))
	copy(newslice, old1)
	copy(newslice[len(old1):], old2)
	return newslice
}

func SetNewReplicaSteps(names []string) {
	newslice := make(MySQLReplicaStepArray, len(names))
	for _, step := range myReplChoosenSteps {
		step.Pos = -1
	}
	for idx, name := range names {
		mySQLReplStepsHash[name].Pos = idx
		newslice = append(newslice, *mySQLReplStepsHash[name])
	}
	myReplChoosenSteps = newslice
}

func RunReplicationStep(name string, test bool) (string, error) {
	if _, ok := mySQLReplStepsHash[name]; ok {
		result, err := mySQLReplStepsHash[name].Run()
		if err != nil {
			return string(result), err
		}
		return string(result), err
	} else {
		err := errors.New("Step wasn't found")
		return "", err
	}
}

/* Backup functions for backup replication steps */

func RestoreStepsListFromBackup(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	var dataRest MySQLReplicaStepArray
	if err := decoder.Decode(&dataRest); err != nil {
		return err
	}
	mySQLReplStepsHash.FromList(dataRest)
	_, selected := mySQLReplStepsHash.GetSelected()
	sort.Sort(selected)
	myReplChoosenSteps = selected
	log.Println(selected)
	return nil
}

func StepsListPrepareForBackup() ([]byte, error) {
	gobBuffer := new(bytes.Buffer)
	gobEnc := gob.NewEncoder(gobBuffer)
	if err := gobEnc.Encode(mySQLReplStepsHash.ToList()); err != nil {
		return nil, err
	}
	return gobBuffer.Bytes(), nil
}

/* end */
