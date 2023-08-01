package k8s_lister

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

const (
	PanicInvalidResourceType = "Same labels for different resources"
)

// K8sListerGroup 多个lister几个,通过字符串进行索引，字符串推荐使用{namespace}{label...}
type (
	K8sListerGroup struct {
		client       *kubernetes.Clientset
		namespace    string
		factory      informers.SharedInformerFactory
		resourceType map[string]K8sResourceType // map[label.String()]K8sResourceType
		mLister      map[string]any             // map[label.String()]k8sLister
	}
)

func NewK8sListerGroup(namespace string, client *kubernetes.Clientset) *K8sListerGroup {
	var k8sListerGroup K8sListerGroup

	k8sListerGroup.namespace = namespace
	k8sListerGroup.client = client
	k8sListerGroup.initFactory()

	k8sListerGroup.mLister = make(map[string]any)
	k8sListerGroup.resourceType = make(map[string]K8sResourceType)

	return &k8sListerGroup
}

// initFactory 初始化factory用于创建不同类型资源的informer
func (klg *K8sListerGroup) initFactory() {
	if klg == nil {
		return
	}

	// 指定namespace的factory
	klg.factory = informers.NewSharedInformerFactoryWithOptions(klg.client, 0, informers.WithNamespace(klg.namespace))
}

// createLister 创建lister
func (klg *K8sListerGroup) createLister(labels labels.Set, resourceType K8sResourceType) (lister any) {
	switch resourceType {
	case ResourcePod:
		lister = newK8sLister[corev1.Pod](klg.factory, ResourcePod, labels)
		return
	case ResourceNode:
		lister = newK8sLister[corev1.Node](klg.factory, ResourceNode, labels)
		return
	case ResourceDeployment:
		lister = newK8sLister[appsv1.Deployment](klg.factory, ResourceDeployment, labels)
		return
	}
	return
}

// ListenPod 监听pod资源
func (klg *K8sListerGroup) ListenPod(
	labels labels.Set,
	onAddHandler func(pod *corev1.Pod),
	onUpdateHandler func(oldPod, newPod *corev1.Pod),
	onDeleteHandler func(pod *corev1.Pod),
	onCntChangedHandler func(cnt int),
) error {
	if klg == nil {
		return fmt.Errorf("K8sListerGroup is nil")
	}

	labelsStr := labels.String()
	klg.resourceType[labelsStr] = ResourcePod
	// 创建lister
	if _, exist := klg.mLister[labelsStr]; !exist {
		klg.mLister[labelsStr] = klg.createLister(labels, ResourcePod)
	} else if klg.mLister[labelsStr] != ResourcePod {
		panic(PanicInvalidResourceType)
	}

	// 注册回调函数
	klg.mLister[labelsStr].(*k8sLister[corev1.Pod]).registerOnAdd(onAddHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Pod]).registerOnUpdate(onUpdateHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Pod]).registerOnDelete(onDeleteHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Pod]).registerOnCntChanged(onCntChangedHandler)

	return nil
}

// ListenNode 监听Node资源
func (klg *K8sListerGroup) ListenNode(
	labels labels.Set,
	onAddHandler func(pod *corev1.Node),
	onUpdateHandler func(oldPod, newPod *corev1.Node),
	onDeleteHandler func(pod *corev1.Node),
	onCntChangedHandler func(cnt int),
) error {
	if klg == nil {
		return fmt.Errorf("K8sListerGroup is nil")
	}

	labelsStr := labels.String()
	klg.resourceType[labelsStr] = ResourceNode
	// 创建lister
	if _, exist := klg.mLister[labelsStr]; !exist {
		klg.mLister[labelsStr] = klg.createLister(labels, ResourceNode)
	} else if klg.mLister[labelsStr] != ResourceNode {
		panic(PanicInvalidResourceType)
	}

	// 注册回调函数
	klg.mLister[labelsStr].(*k8sLister[corev1.Node]).registerOnAdd(onAddHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Node]).registerOnUpdate(onUpdateHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Node]).registerOnDelete(onDeleteHandler)
	klg.mLister[labelsStr].(*k8sLister[corev1.Node]).registerOnCntChanged(onCntChangedHandler)

	return nil
}

// ListenDeployment 监听Deployment资源
func (klg *K8sListerGroup) ListenDeployment(
	labels labels.Set,
	onAddHandler func(pod *appsv1.Deployment),
	onUpdateHandler func(oldPod, newPod *appsv1.Deployment),
	onDeleteHandler func(pod *appsv1.Deployment),
	onCntChangedHandler func(cnt int),
) error {
	if klg == nil {
		return fmt.Errorf("K8sListerGroup is nil")
	}

	labelsStr := labels.String()
	klg.resourceType[labelsStr] = ResourceDeployment
	// 创建lister
	if _, exist := klg.mLister[labelsStr]; !exist {
		klg.mLister[labelsStr] = klg.createLister(labels, ResourceDeployment)
	} else if klg.mLister[labelsStr] != ResourceDeployment {
		panic(PanicInvalidResourceType)
	}

	// 注册回调函数
	klg.mLister[labelsStr].(*k8sLister[appsv1.Deployment]).registerOnAdd(onAddHandler)
	klg.mLister[labelsStr].(*k8sLister[appsv1.Deployment]).registerOnUpdate(onUpdateHandler)
	klg.mLister[labelsStr].(*k8sLister[appsv1.Deployment]).registerOnDelete(onDeleteHandler)
	klg.mLister[labelsStr].(*k8sLister[appsv1.Deployment]).registerOnCntChanged(onCntChangedHandler)

	return nil
}

func (klg *K8sListerGroup) Run(labels labels.Set, stopCh <-chan struct{}) {
	if klg == nil {
		return
	}

	labelsStr := labels.String()
	switch klg.resourceType[labelsStr] {
	case ResourcePod:
		klg.mLister[labelsStr].(*k8sLister[corev1.Pod]).run(stopCh)
		return
	case ResourceNode:
		klg.mLister[labelsStr].(*k8sLister[corev1.Node]).run(stopCh)
		return
	case ResourceDeployment:
		klg.mLister[labelsStr].(*k8sLister[appsv1.Deployment]).run(stopCh)
		return
	}
}
