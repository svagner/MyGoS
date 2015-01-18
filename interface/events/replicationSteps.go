package events

import (
	"encoding/json"

	"github.com/svagner/MyGoS/interface/convert"
	"github.com/svagner/MyGoS/interface/databases"
)

func replicationStepsSubscribe(event string, ip string) {
	replicaSt := make([]string, 0)
	for _, step := range databases.GetChoosenReplicaSteps() {
		replicaSt = append(replicaSt, step.Name)
	}
	Data := ResCmd{Channel: event, Command: "update", Data: replicaSt}
	res, err := json.Marshal(Data)
	for _, user := range Events[event].subscribers {

		if err != nil {
			Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
			user.c <- convert.ConvertToJSON_HTML(Data)
		} else {
			user.c <- string(res)
		}
	}
}

func ReplicationStepAdd(script databases.MySQLReplicaScript) {
	databases.AddReplicationStep(script)
	Data := ResCmd{Channel: "replicationSteps", Command: "add", Data: script}
	res, err := json.Marshal(Data)
	for _, user := range Events["replicationSteps"].subscribers {
		if err != nil {
			Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
			user.c <- convert.ConvertToJSON_HTML(Data)
		} else {
			user.c <- string(res)
		}
	}
}

func ReplicationStepDelete(scriptName string) {
	databases.DeleteReplicationStep(scriptName)
	Data := ResCmd{Channel: "replicationSteps", Command: "delete", Data: scriptName}
	res, err := json.Marshal(Data)
	for _, user := range Events["replicationSteps"].subscribers {
		if err != nil {
			Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
			user.c <- convert.ConvertToJSON_HTML(Data)
		} else {
			user.c <- string(res)
		}
	}
}

func SaveReplicationStepsSelected(co chan string, data []string) {
	databases.SetNewReplicaSteps(data)
	Data := ResCmd{Channel: "replicationSteps", Command: "reinit", Data: data}
	res, err := json.Marshal(Data)
	for _, user := range Events["replicationSteps"].subscribers {
		if err != nil {
			Data := ResCmd{Channel: "Error", Command: "new", Data: err.Error()}
			user.c <- convert.ConvertToJSON_HTML(Data)
		} else {
			if user.c != co {
				user.c <- string(res)
			}
		}
	}
}
