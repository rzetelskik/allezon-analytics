// Copyright 2014-2022 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"strconv"
	"strings"

	"github.com/aerospike/aerospike-client-go/v6/types"
)

// ExecuteTask is used to poll for long running server execute job completion.
type ExecuteTask struct {
	*baseTask

	taskID uint64
	scan   bool
}

// NewExecuteTask initializes task with fields needed to query server nodes.
func NewExecuteTask(cluster *Cluster, statement *Statement) *ExecuteTask {
	return &ExecuteTask{
		baseTask: newTask(cluster),
		taskID:   statement.TaskId,
		scan:     statement.IsScan(),
	}
}

// IsDone queries all nodes for task completion status.
func (etsk *ExecuteTask) IsDone() (bool, Error) {
	var module string
	if etsk.scan {
		module = "scan"
	} else {
		module = "query"
	}

	taskId := strconv.FormatUint(etsk.taskID, 10)

	cmd1 := "query-show:trid=" + taskId
	cmd2 := module + "-show:trid=" + taskId
	cmd3 := "jobs:module=" + module + ";cmd=get-job;trid=" + taskId

	nodes := etsk.cluster.GetNodes()

	for _, node := range nodes {
		var command string
		if node.SupportsPartitionQuery() {
			// query-show works for both scan and query.
			command = cmd1
		} else if node.SupportsQueryShow() {
			command = cmd2
		} else {
			command = cmd3
		}

		responseMap, err := node.requestInfoWithRetry(&etsk.cluster.infoPolicy, 5, command)
		if err != nil {
			return false, err
		}
		response := responseMap[command]

		if strings.HasPrefix(response, "ERROR:2") {
			if etsk.retries.Get() > 20 {
				// Task not found. This could mean task already completed or
				// task not started yet.  We are going to have to assume that
				// the task already completed...
				continue
			}
			// Task not found, and number of retries are not enough; assume task is not started yet.
			return false, nil
		}

		if strings.HasPrefix(response, "ERROR:") {
			// Mark done and quit immediately.
			return false, newError(types.UDF_BAD_RESPONSE, response)
		}

		find := "status="
		index := strings.Index(response, find)

		if index < 0 {
			return false, nil
		}

		begin := index + len(find)
		response = response[begin:]
		find = ":"
		index = strings.Index(response, find)

		if index < 0 {
			continue
		}

		status := strings.ToLower(response[:index])
		if !strings.HasPrefix(status, "done") {
			return false, nil
		}
	}

	return true, nil
}

// OnComplete returns a channel which will be closed when the task is
// completed.
// If an error is encountered while performing the task, an error
// will be sent on the channel.
func (etsk *ExecuteTask) OnComplete() chan Error {
	return etsk.onComplete(etsk)
}