package worker

import (
	"context"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
	"time"
)

// register： /cron/workers/ IP
type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string
}

// initialize the InitRegister
func InitRegister() (err error) {
	var (
		config          clientv3.Config
		client          *clientv3.Client
		localIp         string
		configureSource interface{}
		configure       utils.IniParser
		RegisterSev     Register
	)

	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		return
	}
	configure = configureSource.(utils.IniParser)
	config = clientv3.Config{
		Endpoints:   []string{configure.GetString("etcd", "endpoints")},
		DialTimeout: time.Duration(configure.GetInt64("etcd", "dial_timeout")) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	// get local ip address
	if localIp, err = utils.GetLocalIP(); err != nil {
		return
	}

	// get etcd.KV and etcd.Lease apis
	common.Manage.SetSingleton("Register", Register{
		client:  client,
		kv:      clientv3.NewKV(client),
		lease:   clientv3.NewLease(client),
		localIP: localIp,
	})

	RegisterSev = common.Manage.GetSingleton("JobManager").(Register)
	go RegisterSev.keepOnline()
	return
}

// register /cron/workers
func (register *Register) keepOnline() {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
	)

	for {
		// reset the cancel method
		cancelFunc = nil

		// create grant
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// automatic renewal
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}
		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		// register etcd
		if _, err = register.kv.Put(cancelCtx, common.JOB_WORKER_DIR+register.localIP, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}

		// lease reply
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // if lease fails, try again ！
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}
