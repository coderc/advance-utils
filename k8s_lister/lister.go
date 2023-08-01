package k8s_lister

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type (
	K8sResourceType string
)

const (
	ResourceNode       K8sResourceType = "node"
	ResourcePod        K8sResourceType = "pod"
	ResourceDeployment K8sResourceType = "deployment"
)

const (
	errorInvalidResourceType = "invalid resource type, resource should be: %s"
)

type (
	k8sLister[T corev1.Node | corev1.Pod | appsv1.Deployment] struct {
		t            K8sResourceType
		m            sync.Map // map[string]*T
		cnt          int
		labels       labels.Set
		informer     cache.SharedIndexInformer
		onAdd        func(obj *T)
		onUpdate     func(oldObj, newObj *T)
		onDelete     func(obj *T)
		onCntChanged func(cnt int)
	}
)

// NewK8sLister 创建一个k8sLister, 需要指定资源类型
func newK8sLister[T corev1.Node | corev1.Pod | appsv1.Deployment](factory informers.SharedInformerFactory, resourceType K8sResourceType, labels labels.Set) *k8sLister[T] {
	lister := &k8sLister[T]{}
	lister.t = resourceType
	lister.labels = labels
	lister.initInformer(factory)
	return lister
}

func (kl *k8sLister[T]) initInformer(factory informers.SharedInformerFactory) {
	switch kl.t {
	case ResourcePod:
		kl.informer = factory.Core().V1().Pods().Informer()
		return
	case ResourceDeployment:
		kl.informer = factory.Apps().V1().Deployments().Informer()
		return
	case ResourceNode:
		kl.informer = factory.Core().V1().Nodes().Informer()
		return
	}
}

// getResourceName 获取资源的name
func (kl *k8sLister[T]) getResourceName(obj interface{}) string {
	switch kl.t {
	case ResourceNode:
		return obj.(*corev1.Node).Name
	case ResourcePod:
		return obj.(*corev1.Pod).Name
	case ResourceDeployment:
		return obj.(*appsv1.Deployment).Name
	}

	return ""
}

func (kl *k8sLister[T]) getResourceLabels(obj any) labels.Set {
	switch kl.t {
	case ResourceNode:
		return obj.(*corev1.Node).Labels
	case ResourcePod:
		return obj.(*corev1.Pod).Labels
	case ResourceDeployment:
		return obj.(*appsv1.Deployment).Labels
	}
	return nil
}

// add 资源被添加
func (kl *k8sLister[T]) add(obj *T) {
	name := kl.getResourceName(obj)

	oldObj, ok := kl.m.LoadOrStore(name, obj)

	if ok { // 这个资源之前被添加过
		if kl.onUpdate != nil {
			kl.onUpdate(oldObj.(*T), obj)
		}
	} else { // 新添加的资源
		kl.cnt++
		if kl.onAdd != nil {
			kl.onAdd(obj)
		}
		if kl.onCntChanged != nil {
			kl.onCntChanged(kl.cnt)
		}
	}
}

func (kl *k8sLister[T]) update(oldObj, newObj *T) {
	name := kl.getResourceName(newObj)
	kl.m.Delete(kl.getResourceName(oldObj))
	kl.m.Store(name, newObj)
	if kl.onUpdate != nil {
		kl.onUpdate(oldObj, newObj)
	}
}

// delete 资源被删除
func (kl *k8sLister[T]) delete(obj interface{}) {
	name := kl.getResourceName(obj)
	_, ok := kl.m.LoadAndDelete(name)

	if ok {
		kl.cnt--
		if kl.onDelete != nil {
			kl.onDelete(obj.(*T))
		}
		if kl.onCntChanged != nil {
			kl.onCntChanged(kl.cnt)
		}
	} else { // TODO 删除了一个不存在的资源

	}
}

// Get 通过资源name获取资源
func (kl *k8sLister[T]) get(name string) (*T, bool) {
	obj, ok := kl.m.Load(name)
	if !ok {
		return nil, false
	}
	return obj.(*T), true
}

// List 获取现存的资源列表
func (kl *k8sLister[T]) list() []*T {
	var list []*T
	kl.m.Range(func(key, value interface{}) bool {
		list = append(list, (value).(*T))
		return true
	})
	return list
}

// registerOnAdd 注册资源添加回调
func (kl *k8sLister[T]) registerOnAdd(onAddHandler func(obj *T)) {
	if onAddHandler == nil {
		return
	}
	kl.onAdd = onAddHandler
}

// registerOnUpdate 注册资源更新回调
func (kl *k8sLister[T]) registerOnUpdate(onUpdateHandler func(oldObj, newObj *T)) {
	if onUpdateHandler == nil {
		return
	}
	kl.onUpdate = onUpdateHandler
}

// registerOnDelete 注册资源删除回调
func (kl *k8sLister[T]) registerOnDelete(onDeleteHandler func(obj *T)) {
	if onDeleteHandler == nil {
		return
	}
	kl.onDelete = onDeleteHandler
}

// registerOnCntChanged 注册资源数量变化回调
func (kl *k8sLister[T]) registerOnCntChanged(onCntChangedHandler func(cnt int)) {
	if onCntChangedHandler == nil {
		return
	}
	kl.onCntChanged = onCntChangedHandler
}

// bind 绑定事件函数
func (kl *k8sLister[T]) bind() error {
	_, err := kl.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			obj2T := obj.(*T)
			if kl.filterResource(obj2T) {
				kl.add(obj2T)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldObj2T := oldObj.(*T)
			newObj2T := newObj.(*T)
			oldOk, newOk := kl.filterResource(oldObj2T), kl.filterResource(newObj2T)
			if oldOk && newOk { // 更新
				kl.update(oldObj2T, newObj2T)
			} else if oldOk { // 删除
				kl.delete(oldObj2T)
			} else if newOk { // 添加
				kl.add(newObj2T)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newLabels := kl.getResourceLabels(obj)
			if kl.labels.AsSelector().Matches(newLabels) {
				kl.delete(obj.(*T))
			}
		},
	})

	return err
}

func (kl *k8sLister[T]) filterResource(obj *T) bool {
	return kl.labels.AsSelector().Matches(kl.getResourceLabels(obj))
}

// Run 启动informer
func (kl *k8sLister[T]) run(stopCh <-chan struct{}) {
	err := kl.bind()
	if err != nil {
		panic(err)
	}

	kl.informer.Run(stopCh)
}
