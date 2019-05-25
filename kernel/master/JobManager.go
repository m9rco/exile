package master

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/m9rco/exile/_vendor-20190511225511/github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
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
		return
	}
	configure := configureSource.(utils.IniParser)
	config = clientv3.Config{
		Endpoints:            []string{configure.GetString("etcd", "endpoints")},
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(configure.GetInt64("etcd", "dial_timeout")) * time.Millisecond,
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

func (jobMgr *JobManager) ListJobs() (jobList []*common.Job, err error) {
	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	fmt.Println(getResp)
	jobList = make([]*common.Job, 0)
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
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

func (jobMgr *JobManager) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey     string
		deleteResp *clientv3.DeleteResponse
		oldJobObj  common.Job
	)
	jobKey = common.JOB_SAVE_DIR + name
	if deleteResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	if len(deleteResp.PrevKvs) != 0 {
		if err = json.Unmarshal(deleteResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (jobMgr *JobManager) KillJob(jobName string) (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
	)
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}
	_, err = jobMgr.kv.Put(context.TODO(), common.JOB_KILLER_DIR+jobName, "", clientv3.WithLease(leaseGrantResp.ID))
	return
}
