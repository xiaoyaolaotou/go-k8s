package main

import (
	"context"
	"fmt"
	"kubernetes/testk8s/utils"

	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"log"
	"time"
)

func main() {
	kubeconfig := "/root/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	// 1. namespace列表获取
	namespaceClient := clientset.CoreV1().Namespaces()
	namespaceResult, err := namespaceClient.List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	namespaces := []string{}
	fmt.Println("namespaces:")
	for _, namespace := range namespaceResult.Items {
		namespaces = append(namespaces, namespace.Name)
		fmt.Println(namespace.Name, now.Sub(namespace.CreationTimestamp.Time))
	}

	// 2. deployment 列表
	for _, namespace := range namespaces {
		deploymentClient := clientset.ExtensionsV1beta1().Deployments(namespace)
		deploymentResult, err := deploymentClient.List(context.TODO(), metaV1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, deployment := range deploymentResult.Items {
			fmt.Println(deployment.Namespace, deployment.Name)
		}

	}

	// 3. service 列表
	fmt.Println("=======service=======")
	for _, namespace := range namespaces {
		serviceClient := clientset.CoreV1().Services(namespace)
		serviceResult, err := serviceClient.List(context.TODO(), metaV1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, service := range serviceResult.Items {
			fmt.Println(service.Namespace, service.Name, service.Spec.Ports)
		}
	}

	// 4. deployment创建
	deploymentClient := clientset.ExtensionsV1beta1().Deployments("default")
	deployment := &appsv1.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Name: "test-nginx-dev",
			Labels: map[string]string{
				"source": "cmdb",
				"app":    "nginx",
				"env":    "test",
			},
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: utils.Int32Ptr(3),
			Selector: &metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"source": "cmdb",
					"app":    "nginx",
					"env":    "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metaV1.ObjectMeta{
					Labels: map[string]string{
						"source": "cmdb",
						"app":    "nginx",
						"env":    "test",
					},
				},
				//Spec:corev1.PodSpec{
				Spec: corev1.PodSpec{
					//Containers: []corev1.Container{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	//deployment, err = deploymentClient.Create(context.TODO(), deployment, metaV1.CreateOptions{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	fmt.Println(deployment.Status)

	// 5. deployment 修改
	deployment, err = deploymentClient.Get(context.TODO(), "test-nginx-dev", metaV1.GetOptions{})
	if *deployment.Spec.Replicas > 3 {
		deployment.Spec.Replicas = utils.Int32Ptr(1)
	} else {
		deployment.Spec.Replicas = utils.Int32Ptr(*deployment.Spec.Replicas + 1)
	}

	deployment, err = deploymentClient.Update(context.TODO(), deployment, metaV1.UpdateOptions{})
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(deployment.Status)
	}

	// 6. 删除
	deploymentClient.Delete(context.TODO(), "test-nginx-dev", metaV1.DeleteOptions{})





}
