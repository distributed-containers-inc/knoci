package controller

import (
	"fmt"
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"time"
)

type TestListWatcher struct {
	TestsCli *versioned.Clientset

	AddFunc func(test *v1alpha1.Test)
	DeleteFunc func(test *v1alpha1.Test)
	UpdateFunc func(oldTest, newTest *v1alpha1.Test)
}

func (watcher *TestListWatcher) Run() {
	watchlist := cache.NewListWatchFromClient(watcher.TestsCli.RESTClient(), "tests", "", fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1alpha1.Test{},
		time.Duration(0),
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				test := obj.(*v1alpha1.Test)
				fmt.Println("Got a new test: "+test.Name)
				if watcher.AddFunc != nil {
					watcher.AddFunc(test)
				}
			},
			DeleteFunc: func(obj interface{}) {
				test := obj.(*v1alpha1.Test)
				fmt.Println("Deleted test: "+test.Name)
				if watcher.DeleteFunc != nil {
					watcher.DeleteFunc(test)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldTest := oldObj.(*v1alpha1.Test)
				newTest := newObj.(*v1alpha1.Test)
				fmt.Println("Got a changed test: "+oldTest.Name+" -> "+newTest.Name)
				if watcher.UpdateFunc != nil {
					watcher.UpdateFunc(oldTest, newTest)
				}
			},
		},
	)

	stop := make(chan os.Signal, 1)
	stopWrapped := make(chan struct{})
	go func() {
		<-stop
		stopWrapped <- struct{}{}
	}()
	signal.Notify(stop, os.Interrupt)
	controller.Run(stopWrapped)
}
