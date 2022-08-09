package executor

import (
	"context"
	"strconv"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

const (
	_ResourceQuotaName  = "zero-quota"
	_ReplicasAnnotation = apis.AnnotationPrefix + "/restore-replicas"
	_WaitPodsTimeout    = time.Second * 180
	_WaitPodsInterval   = time.Second * 15
)

func (ex *Executor) executeShutdownKube(policy *apis.StandSchedulePolicy) error {
	namespaces, err := ex.fetchNamespaces(policy.Spec.TargetNamespaceFilter, true)
	if err != nil {
		return err
	}

	return util.ForEachE(namespaces, func(_ int, namespace string) error {
		return multierr.Combine(
			ex.scaleDownApps(namespace),
			ex.createResourceQuota(namespace, policy),
			ex.cleanupPods(namespace),
		)
	})
}

func (ex *Executor) executeStartupKube(policy *apis.StandSchedulePolicy) error {
	namespaces, err := ex.fetchNamespaces(policy.Spec.TargetNamespaceFilter, false)
	if err != nil {
		return err
	}

	return util.ForEachE(namespaces, func(_ int, namespace string) error {
		return multierr.Combine(
			ex.deleteResourceQuota(namespace),
			ex.scaleUpApps(namespace),
		)
	})
}

func (ex *Executor) createResourceQuota(namespace string, policy *apis.StandSchedulePolicy) error {
	ex.logger.Debug("Create resource quota in namespace",
		zap.String("quota", _ResourceQuotaName),
		zap.String("namespace", namespace))

	quota := &core.ResourceQuota{
		ObjectMeta: meta.ObjectMeta{
			Name:      _ResourceQuotaName,
			Namespace: namespace,
			OwnerReferences: []meta.OwnerReference{
				{
					APIVersion: apis.GroupVersion.String(),
					Kind:       "StandSchedulePolicy",
					Name:       policy.Name,
					UID:        policy.UID,
				},
			},
		},
		Spec: core.ResourceQuotaSpec{
			Hard: core.ResourceList{
				core.ResourcePods: resource.MustParse("0"),
			},
		},
	}

	_, err := ex.kube.CoreClient().
		CoreV1().
		ResourceQuotas(namespace).
		Create(context.Background(), quota, meta.CreateOptions{})

	return kubernetes.IgnoreAlreadyExistsError(err)
}

func (ex *Executor) scaleDownApps(namespace string) error {
	ex.logger.Debug("ScaleDown deployments and statefulSets in namespace", zap.String("namespace", namespace))

	deployments, err := ex.lister.Deployments.Deployments(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	deployments = util.Where(deployments, func(_ int, d *apps.Deployment) bool {
		return d.Spec.Replicas != nil && *d.Spec.Replicas != 0
	})

	statefulSets, err := ex.lister.StatefulSets.StatefulSets(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	statefulSets = util.Where(statefulSets, func(_ int, s *apps.StatefulSet) bool {
		return s.Spec.Replicas != nil && *s.Spec.Replicas != 0
	})

	return multierr.Combine(
		util.ForEachE(deployments, func(_ int, deployment *apps.Deployment) error {
			replicas := *deployment.Spec.Replicas
			deployment.Spec.Replicas = util.Pointer(int32(0))
			deployment.ObjectMeta.Annotations[_ReplicasAnnotation] = strconv.Itoa(int(replicas))
			return ex.updateDeployment(deployment)
		}),
		util.ForEachE(statefulSets, func(_ int, sts *apps.StatefulSet) error {
			replicas := *sts.Spec.Replicas
			sts.Spec.Replicas = util.Pointer(int32(0))
			sts.ObjectMeta.Annotations[_ReplicasAnnotation] = strconv.Itoa(int(replicas))
			return ex.updateStatefulSet(sts)
		}),
	)
}

func (ex *Executor) cleanupPods(namespace string) error {
	ex.logger.Debug("Delete all existing pods in namespace", zap.String("namespace", namespace))

	return ex.kube.CoreClient().
		CoreV1().
		Pods(namespace).
		DeleteCollection(
			context.Background(),
			meta.DeleteOptions{
				PropagationPolicy: util.Pointer(meta.DeletePropagationBackground),
			},
			meta.ListOptions{},
		)
}

func (ex *Executor) deleteResourceQuota(namespace string) error {
	ex.logger.Debug("Delete resource quota in namespace",
		zap.String("quota", _ResourceQuotaName),
		zap.String("namespace", namespace))

	err := ex.kube.CoreClient().
		CoreV1().
		ResourceQuotas(namespace).
		Delete(context.Background(), _ResourceQuotaName, meta.DeleteOptions{})

	return kubernetes.IgnoreNotFoundError(err)
}

func (ex *Executor) scaleUpApps(namespace string) error {
	ex.logger.Debug("ScaleUp deployments and statefulSets in namespace", zap.String("namespace", namespace))

	deployments, err := ex.lister.Deployments.Deployments(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	deployments = util.Where(deployments, func(_ int, d *apps.Deployment) bool {
		replicas := d.Spec.Replicas != nil && *d.Spec.Replicas == 0
		_, scaled := d.Annotations[_ReplicasAnnotation]
		return replicas && scaled
	})

	statefulSets, err := ex.lister.StatefulSets.StatefulSets(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	statefulSets = util.Where(statefulSets, func(_ int, s *apps.StatefulSet) bool {
		replicas := s.Spec.Replicas != nil && *s.Spec.Replicas == 0
		_, scaled := s.Annotations[_ReplicasAnnotation]
		return replicas && scaled
	})

	return multierr.Combine(
		util.ForEachE(statefulSets, func(_ int, sts *apps.StatefulSet) error {
			replicas, _ := strconv.Atoi(sts.ObjectMeta.Annotations[_ReplicasAnnotation])
			sts.Spec.Replicas = util.Pointer(int32(replicas))
			delete(sts.ObjectMeta.Annotations, _ReplicasAnnotation)
			return ex.updateStatefulSet(sts)
		}),
		ex.waitPods(namespace),
		util.ForEachE(deployments, func(_ int, deployment *apps.Deployment) error {
			replicas, _ := strconv.Atoi(deployment.ObjectMeta.Annotations[_ReplicasAnnotation])
			deployment.Spec.Replicas = util.Pointer(int32(replicas))
			delete(deployment.ObjectMeta.Annotations, _ReplicasAnnotation)
			return ex.updateDeployment(deployment)
		}),
	)
}

func (ex *Executor) waitPods(namespace string) error {
	return wait.Poll(_WaitPodsInterval, _WaitPodsTimeout, func() (bool, error) {
		ex.logger.Debug("Wait pods in namespace", zap.String("namespace", namespace))

		podList, err := ex.kube.CoreClient().
			CoreV1().
			Pods(namespace).
			List(context.Background(), meta.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == core.PodPending {
				return false, err
			}
		}

		return true, nil
	})
}

func (ex *Executor) updateDeployment(deployment *apps.Deployment) error {
	_, err := ex.kube.CoreClient().
		AppsV1().
		Deployments(deployment.Namespace).
		Update(context.Background(), deployment, meta.UpdateOptions{})
	return err
}

func (ex *Executor) updateStatefulSet(sts *apps.StatefulSet) error {
	_, err := ex.kube.CoreClient().
		AppsV1().
		StatefulSets(sts.Namespace).
		Update(context.Background(), sts, meta.UpdateOptions{})
	return err
}

func (ex *Executor) fetchNamespaces(filter string, reverse bool) ([]string, error) {
	list, err := ex.lister.Namespaces.List(labels.Everything())
	if err != nil {
		ex.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return nil, err
	}
	return SortNamespaces(list, filter, reverse), nil
}
