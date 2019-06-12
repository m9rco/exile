WorkerManager.go
package master

import (
  "context"
  "github.com/m9rco/exile/kernel/common"
  "github.com/m9rco/exile/kernel/utils"
  "github.com/coreos/etcd/mvcc/mvccpb"
  "github.com/etcd-io/etcd/clientv3"
  "strings"
  "time"
)

type WorkerManager struct {
  client *clientv3.Client
  kv     clientv3.KV
  lease  clientv3.Lease
}

// get worker list
func (workerMgr *WorkerManager) ListWorkers() (workerArr []string, err error) {
  var (
    getResp  *clientv3.GetResponse
    kv       *mvccpb.KeyValue
    workerIP string
  )
  // new array
  workerArr = make([]string, 0)
  // get all kv
  if getResp, err = workerMgr.kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix()); err != nil {
    return
  }

  // parsing IP of each node
  for _, kv = range getResp.Kvs {
    workerIP = common.ExtractWorkerIP(string(kv.Key))
    workerArr = append(workerArr, workerIP)
  }
  return
}

func InitWorkerMgr() (err error) {
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
    Endpoints:   strings.Split(configure.GetString("etcd", "endpoints"), ","),
    DialTimeout: time.Duration(configure.GetInt64("etcd", "dial_timeout")) * time.Millisecond,
  }

  // build connect
  if client, err = clientv3.New(config); err != nil {
    return
  }

  // get KV and Lease API child
  common.Manage.SetSingleton("WorkerManager", WorkerManager{
    client: client,
    kv:     clientv3.NewKV(client),
    lease:  clientv3.NewLease(client),
  })

  return
}
