package controller

import (
	"github.com/distributed-containers-inc/knoci/pkg/apis/testing/v1alpha1"
	"github.com/distributed-containers-inc/knoci/pkg/client/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
	"time"
)

type TestListWatcher struct {
	TestsCli *versioned.Clientset

	AddFunc    func(test *v1alpha1.Test)
	DeleteFunc func(test *v1alpha1.Test)
	UpdateFunc func(oldTest, newTest *v1alpha1.Test)
}

func (watcher *TestListWatcher) Run() {
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (object runtime.Object, e error) {
				return watcher.TestsCli.TestingV1alpha1().Tests("").List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (i watch.Interface, e error) {
				return watcher.TestsCli.TestingV1alpha1().Tests("").Watch(options)
			},
		},
		&v1alpha1.Test{},
		time.Duration(0),
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				test := obj.(*v1alpha1.Test)
				if watcher.AddFunc != nil {
					watcher.AddFunc(test)
				}
			},
			DeleteFunc: func(obj interface{}) {
				test := obj.(*v1alpha1.Test)
				if watcher.DeleteFunc != nil {
					watcher.DeleteFunc(test)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldTest := oldObj.(*v1alpha1.Test)
				newTest := newObj.(*v1alpha1.Test)
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
