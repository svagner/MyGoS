package events

import (
	"encoding/json"

	"github.com/svagner/MyGoS/interface/convert"
	"github.com/svagner/MyGoS/interface/databases"
)

type ResCmd struct {
	Channel string
	Command string
	Data    interface{}
}

func ReplicationGroups(data string, co chan string, ip string) error {
	if err := databases.AddReplicaGroup(data); err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	Data := ResCmd{Channel: "replicationGroups", Command: "add", Data: data}
	Events["replicationGroups"].channel <- convert.ConvertToJSON_HTML(Data)
	return nil
}

func ReplicationGroupsDelete(data string, co chan string, ip string) error {
	databases.DeleteReplicaGroup(data)
	Data := ResCmd{Channel: "replicationGroups", Command: "delete", Data: data}
	Events["replicationGroups"].channel <- convert.ConvertToJSON_HTML(Data)
	return nil
}

func ReplicationGroupsEdit(data string, co chan string, ip string) error {
	var RenameData struct {
		From string
		To   string
	}
	if err := json.Unmarshal([]byte(data), &RenameData); err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	if err := databases.EditReplicaGroup(RenameData.From, RenameData.To); err != nil {
		Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
		co <- convert.ConvertToJSON_HTML(Data)
		return err
	}
	Data := ResCmd{Channel: "replicationGroups", Command: "update", Data: data}
	Events["replicationGroups"].channel <- convert.ConvertToJSON_HTML(Data)
	return nil
}
