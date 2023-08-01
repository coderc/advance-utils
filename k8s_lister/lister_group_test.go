package k8s_lister

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"testing"
	"time"
)

const (
	TestNamespace = "dsp-ns"
)

func TestListerGroup(t *testing.T) {
	listerGroup := NewK8sListerGroup(TestNamespace, nil)
	t.Run("Input=test listen pods valid", func(t *testing.T) {
		labelsMatch := labels.Set{"run": "api-server"}
		err := listerGroup.ListenPod(
			labelsMatch,
			func(pod *corev1.Pod) {
				fmt.Printf("add pod:\n%v\n", pod)
			},
			func(oldPod, newPod *corev1.Pod) {
				fmt.Printf("update pod old pod is:\n%v\n", oldPod)
				fmt.Printf("update pod new pod is:\n%v\n", newPod)
			},
			func(pod *corev1.Pod) {
				fmt.Printf("delete pod:\n%v\n", pod)
			},
			func(cnt int) {
				fmt.Printf("cnt changed to %d\n", cnt)
			},
		)

		assert.NoError(t, err)

		stopCh := make(chan struct{})
		go listerGroup.Run(labelsMatch, stopCh)

		// goroutine被杀死时，停止pod监听
		defer close(stopCh)

		// 监听三十秒内pod的变化
		timeout := time.After(time.Second * 10)
		select {
		case <-timeout:
			return
		}
	})

	t.Run("Input=test listen node valid", func(t *testing.T) {
		labelsMatch := labels.Set{"run": "cost-server"}
		err := listerGroup.ListenNode(
			labelsMatch,
			func(node *corev1.Node) {
				fmt.Printf("add node:\n%v\n", node)
			},
			func(oldNode, newNode *corev1.Node) {
				fmt.Printf("update node old node is:\n%v\n", oldNode)
				fmt.Printf("update node new node is:\n%v\n", newNode)
			},
			func(node *corev1.Node) {
				fmt.Printf("delete node:\n%v\n", node)
			},
			func(cnt int) {
				fmt.Printf("cnt changed to %d\n", cnt)
			},
		)

		assert.NoError(t, err)

		stopCh := make(chan struct{})
		go listerGroup.Run(labelsMatch, stopCh)

		// goroutine被杀死时，停止node监听
		defer close(stopCh)

		// 监听三十秒内node的变化
		timeout := time.After(time.Second * 10)
		select {
		case <-timeout:
			return
		}
	})
}
