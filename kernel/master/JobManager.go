package master

import (
	"encoding/json"
	"fmt"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/m9rco/exile/_vendor-20190511225511/github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
	"os"
	"time"
)

type JobManager struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func InitJobMgr() (err error) {
	var (
		config          clientv3.Config
		client          *clientv3.Client
		configureSource interface{}
	)
	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		fmt.Printf("fail to read file: %v", err)
		os.Exit(1)
	}
	configure := configureSource.(utils.IniParser)
	config = clientv3.Config{
		Endpoints:            []string{configure.GetString("server", "etcd_endpoints")},
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(configure.GetInt64("server", "etcd_timeout")) * time.Millisecond,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}

	common.Manage.SetSingleton("JobManager", JobManager{
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  clientv3.NewLease(client),
	})

	return
}

func (jobMgr *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	jobKey = common.JOB_SAVE_DIR + job.Name
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	if putResp.PrevKv != nil {
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
