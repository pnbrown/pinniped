// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"go.pinniped.dev/test/library"
)

const (
	kubeCertAgentLabelSelector = "kube-cert-agent.pinniped.dev=true"
)

func TestKubeCertAgent(t *testing.T) {
	env := library.IntegrationEnv(t).WithCapability(library.ClusterSigningKeyIsAvailable)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	kubeClient := library.NewClientset(t)

	// Make sure the agent pods are running and healthy before the tests begin.
	t.Logf("waiting for agent pods to become running before tests")
	waitForAllAgentsRunning(ctx, t, kubeClient, env)

	// Get the current number of kube-cert-agent pods.
	//
	// We can pretty safely assert there should be more than 1, since there should be a
	// kube-cert-agent pod per kube-controller-manager pod, and there should probably be at least
	// 1 kube-controller-manager for this to be a working kube API.
	originalAgentPods, err := kubeClient.CoreV1().Pods(env.ConciergeNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: kubeCertAgentLabelSelector,
	})
	require.NoError(t, err)
	require.NotEmpty(t, originalAgentPods.Items)
	sortPods(originalAgentPods)

	for _, agentPod := range originalAgentPods.Items {
		// All agent pods should contain all custom labels
		for k, v := range env.ConciergeCustomLabels {
			require.Equalf(t, v, agentPod.Labels[k], "expected agent pod to have label `%s: %s`", k, v)
		}
		require.Equal(t, env.ConciergeAppName, agentPod.Labels["app"])
	}

	agentPodsReconciled := func() bool {
		var currentAgentPods *corev1.PodList
		currentAgentPods, err = kubeClient.CoreV1().Pods(env.ConciergeNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: kubeCertAgentLabelSelector,
		})

		if err != nil {
			return false
		}

		if len(originalAgentPods.Items) != len(currentAgentPods.Items) {
			err = fmt.Errorf(
				"original agent pod len != current agent pod len: %s",
				diff.ObjectDiff(originalAgentPods.Items, currentAgentPods.Items),
			)
			return false
		}

		sortPods(currentAgentPods)
		for i := range originalAgentPods.Items {
			if !equality.Semantic.DeepEqual(
				originalAgentPods.Items[i].Spec,
				currentAgentPods.Items[i].Spec,
			) {
				err = fmt.Errorf(
					"original agent pod != current agent pod: %s",
					diff.ObjectDiff(originalAgentPods.Items[i].Spec, currentAgentPods.Items[i].Spec),
				)
				return false
			}
		}

		return true
	}

	t.Run("reconcile on update", func(t *testing.T) {
		// Update the image of the first pod. The controller should see it, and flip it back.
		//
		// Note that we update the toleration field here because it is the only field, currently, that
		// 1) we are allowed to update on a running pod AND 2) the kube-cert-agent controllers care
		// about.
		updatedAgentPod := originalAgentPods.Items[0].DeepCopy()
		updatedAgentPod.Spec.Tolerations = append(
			updatedAgentPod.Spec.Tolerations,
			corev1.Toleration{Key: "fake-toleration"},
		)
		t.Logf("updating agent pod %s/%s with a fake toleration", updatedAgentPod.Namespace, updatedAgentPod.Name)
		_, err = kubeClient.CoreV1().Pods(env.ConciergeNamespace).Update(ctx, updatedAgentPod, metav1.UpdateOptions{})
		require.NoError(t, err)
		time.Sleep(1 * time.Second)

		// Make sure the original pods come back.
		t.Logf("waiting for agent pods to reconcile")
		assert.Eventually(t, agentPodsReconciled, 10*time.Second, 250*time.Millisecond)
		require.NoError(t, err)

		// Make sure the pods all become healthy.
		t.Logf("waiting for agent pods to become running")
		waitForAllAgentsRunning(ctx, t, kubeClient, env)
	})

	t.Run("reconcile on delete", func(t *testing.T) {
		// Delete the first pod. The controller should see it, and flip it back.
		podToDelete := originalAgentPods.Items[0]
		t.Logf("deleting agent pod %s/%s", podToDelete.Namespace, podToDelete.Name)
		err = kubeClient.
			CoreV1().
			Pods(env.ConciergeNamespace).
			Delete(ctx, podToDelete.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		time.Sleep(1 * time.Second)

		// Make sure the original pods come back.
		t.Logf("waiting for agent pods to reconcile")
		assert.Eventually(t, agentPodsReconciled, 10*time.Second, 250*time.Millisecond)
		require.NoError(t, err)

		// Make sure the pods all become healthy.
		t.Logf("waiting for agent pods to become running")
		waitForAllAgentsRunning(ctx, t, kubeClient, env)
	})

	t.Run("reconcile on unhealthy", func(t *testing.T) {
		// Refresh this pod so we have its latest UID to compare against later.
		podToDisrupt := &originalAgentPods.Items[0]
		podToDisrupt, err = kubeClient.CoreV1().Pods(podToDisrupt.Namespace).Get(ctx, originalAgentPods.Items[0].Name, metav1.GetOptions{})
		require.NoError(t, err)

		// Exec into the pod and kill the sleep process, which should cause the pod to enter status.phase == "Error".
		execRequest := kubeClient.
			CoreV1().
			RESTClient().
			Post().
			Namespace(podToDisrupt.Namespace).
			Resource("pods").
			Name(podToDisrupt.Name).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Stdout:  true,
				Stderr:  true,
				Command: []string{"/usr/bin/killall", "sleep"},
			}, scheme.ParameterCodec)
		executor, err := remotecommand.NewSPDYExecutor(library.NewClientConfig(t), "POST", execRequest.URL())
		require.NoError(t, err)
		t.Logf("execing into agent pod %s/%s to run '/usr/bin/killall sleep'", podToDisrupt.Namespace, podToDisrupt.Name)
		var stdout, stderr bytes.Buffer
		err = executor.Stream(remotecommand.StreamOptions{Stdout: &stdout, Stderr: &stderr})

		// Some container runtimes (e.g., in CI) exit fast enough that our killall process also gets a SIGKILL.
		if err != nil && strings.Contains(err.Error(), "command terminated with exit code 137") {
			t.Logf("ignoring SIGKILL error: %s", err.Error())
			err = nil
		}

		require.NoError(t, err)
		t.Logf("'/usr/bin/killall sleep' finished (stdout: %q, stderr: %q)", stdout.String(), stderr.String())

		// Wait for that pod to be disappear (since it will have failed).
		t.Logf("waiting for unhealthy agent pod to disappear")
		library.RequireEventuallyWithoutError(t, func() (bool, error) {
			currentPod, err := kubeClient.CoreV1().Pods(podToDisrupt.Namespace).Get(ctx, podToDisrupt.Name, metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			if currentPod.UID == podToDisrupt.UID {
				t.Logf("pod %s/%s still exists in status %s", podToDisrupt.Namespace, podToDisrupt.Name, currentPod.Status.Phase)
				return false, nil
			}
			return true, nil
		}, 10*time.Second, 1*time.Second, "unhealthy agent pod was never deleted")

		t.Logf("waiting for agent pods to reconcile")
		// Make sure the original pods come back.
		assert.Eventually(t, agentPodsReconciled, 10*time.Second, 250*time.Millisecond)
		require.NoError(t, err)

		// Make sure the pods all become healthy.
		t.Logf("waiting for agent pods to become running")
		waitForAllAgentsRunning(ctx, t, kubeClient, env)
	})
}

func waitForAllAgentsRunning(ctx context.Context, t *testing.T, kubeClient kubernetes.Interface, env *library.TestEnv) {
	library.RequireEventuallyWithoutError(t, func() (bool, error) {
		pods, err := kubeClient.CoreV1().Pods(env.ConciergeNamespace).List(ctx, metav1.ListOptions{
			LabelSelector: kubeCertAgentLabelSelector,
		})
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			t.Logf("there are no agent pods yet")
			return false, nil
		}

		allRunning := true
		for _, pod := range pods.Items {
			t.Logf("agent pod %s/%s is in status %s", pod.Namespace, pod.Name, pod.Status.Phase)
			if pod.Status.Phase != corev1.PodRunning {
				allRunning = false
			}
		}
		return allRunning, nil
	}, 60*time.Second, 2*time.Second, "agent pods never went back to Running status")
}

func sortPods(pods *corev1.PodList) {
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].Name < pods.Items[j].Name
	})
}
