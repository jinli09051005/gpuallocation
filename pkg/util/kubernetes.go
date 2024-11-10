package util

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// 为node节点添加显存容量注解
func UpdateCurrentNode(ctx context.Context, nodeName, gpumemes string) error {
	k8sClient := getClient()
	node, err := k8sClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	// jinli.io/gpumems=uuid1_1024,uuid2_2048
	node.Annotations["jinli.io/gpumems"] = gpumemes

	_, err = k8sClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return err
	} else {
		klog.Infoln("Node annotations updated successfully.")
	}
	return nil
}

// 获取处于allocating状态Pod
func GetCurrentPod(ctx context.Context, nodeName string) (*corev1.Pod, error) {
	selector := fmt.Sprintf("spec.nodeName=%s", nodeName)
	podListOptions := metav1.ListOptions{
		FieldSelector: selector,
	}

	pods, err := getClient().CoreV1().Pods("").List(ctx, podListOptions)
	if err != nil {
		return nil, err
	}
	for _, p := range pods.Items {
		if p.Status.Phase != corev1.PodPending {
			continue
		}

		if status, ok := p.Annotations["AllocateStatus"]; !ok {
			continue
		} else {
			if status == "allocating" {
				return &p, nil
			}
		}
	}
	return nil, fmt.Errorf("no binding pod found on node %s", nodeName)
}

// 更新pod分配状态
func UpdateCurrentPod(ctx context.Context, pod *corev1.Pod) error {
	k8sClient := getClient()
	annotations := make(map[string]string)
	// pod注解
	// "AllocateStatus":"allocated"
	// "name":"uuid"
	for _, v := range pod.Spec.Containers {
		var value string
		for _, env := range v.Env {
			if env.Name == "UUID" {
				value = env.Value
			}
		}
		annotations[v.Name] = value
	}
	annotations["AllocateStatus"] = "allocated"
	patchData := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": annotations,
		},
	}
	patchDataByte, _ := json.Marshal(patchData)
	// 更新Pod注解
	_, err := k8sClient.CoreV1().Pods(pod.Namespace).Patch(ctx, pod.Name, types.StrategicMergePatchType, patchDataByte, metav1.PatchOptions{})
	if err != nil {
		return err
	} else {
		klog.Infoln("Pod annotations updated successfully.")
	}
	return nil
}

// func UpdateCurrentPod(ctx context.Context, podName, podNs string) error {
// 	k8sClient := getClient()
// 	pod, err := k8sClient.CoreV1().Pods(podNs).Get(ctx, podName, metav1.GetOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	if pod.Annotations == nil {
// 		pod.Annotations = make(map[string]string)
// 	}

// 	pod.Annotations["AllocateStatus"] = "allocated"

// 	_, err = k8sClient.CoreV1().Pods(podNs).Update(context.TODO(), pod, metav1.UpdateOptions{})
// 	if err != nil {
// 		return err
// 	} else {
// 		klog.Infoln("Pod annotations updated successfully.")
// 	}
// 	return nil
// }

func getClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("InClusterConfig Failed for err:%s", err.Error())
		panic(err)
	}
	KubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("new config error %s", err.Error())
		panic(err)
	}
	return KubeClient
}
