package executor

import (
	"context"
	"strconv"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

const (
	_ResourceQuotaName  = "zero-quota"
	_ReplicasAnnotation = apis.AnnotationPrefix + "/restore-replicas"
)

// todo: wait pods between scales

func (in *Executor) executeShutdownKube(policy *apis.StandSchedulePolicy) error {
	namespaces, err := in.fetchNamespaces(policy.Spec.TargetNamespaceFilter, true)
	if err != nil {
		in.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	return util.ForEachE(namespaces, func(_ int, namespace string) error {
		return multierr.Combine(
			in.scaleDownApps(namespace),
			in.createResourceQuota(namespace, policy),
			in.cleanupPods(namespace),
		)
	})
}

func (in *Executor) executeStartupKube(policy *apis.StandSchedulePolicy) error {
	namespaces, err := in.fetchNamespaces(policy.Spec.TargetNamespaceFilter, false)
	if err != nil {
		in.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	return util.ForEachE(namespaces, func(_ int, namespace string) error {
		return multierr.Combine(
			in.deleteResourceQuota(namespace),
			in.scaleUpApps(namespace),
		)
	})
}

func (in *Executor) createResourceQuota(namespace string, policy *apis.StandSchedulePolicy) error {
	in.logger.Debug("Create resource quota in namespace",
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

	_, err := in.kube.CoreClient().
		CoreV1().
		ResourceQuotas(namespace).
		Create(context.Background(), quota, meta.CreateOptions{})

	return kubernetes.IgnoreAlreadyExistsError(err)
}

func (in *Executor) scaleDownApps(namespace string) error {
	in.logger.Debug("ScaleDown deployments and statefulSets in namespace",
		zap.String("namespace", namespace))

	deployments, err := in.lister.Deployments.Deployments(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	deployments = util.Where(deployments, func(_ int, d *apps.Deployment) bool {
		return d.Spec.Replicas != nil && *d.Spec.Replicas != 0
	})

	statefulSets, err := in.lister.StatefulSets.StatefulSets(namespace).List(labels.Everything())
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
			return in.updateDeployment(deployment)
		}),
		util.ForEachE(statefulSets, func(_ int, sts *apps.StatefulSet) error {
			replicas := *sts.Spec.Replicas
			sts.Spec.Replicas = util.Pointer(int32(0))
			sts.ObjectMeta.Annotations[_ReplicasAnnotation] = strconv.Itoa(int(replicas))
			return in.updateStatefulSet(sts)
		}),
	)
}

func (in *Executor) cleanupPods(namespace string) error {
	in.logger.Debug("Delete all existing pods in namespace",
		zap.String("namespace", namespace))

	return in.kube.CoreClient().
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

func (in *Executor) deleteResourceQuota(namespace string) error {
	in.logger.Debug("Delete resource quota in namespace",
		zap.String("quota", _ResourceQuotaName),
		zap.String("namespace", namespace))

	err := in.kube.CoreClient().
		CoreV1().
		ResourceQuotas(namespace).
		Delete(context.Background(), _ResourceQuotaName, meta.DeleteOptions{})

	return kubernetes.IgnoreNotFoundError(err)
}

func (in *Executor) scaleUpApps(namespace string) error {
	in.logger.Debug("ScaleUp deployments and statefulSets in namespace",
		zap.String("namespace", namespace))

	deployments, err := in.lister.Deployments.Deployments(namespace).List(labels.Everything())
	if err != nil {
		return err
	}
	deployments = util.Where(deployments, func(_ int, d *apps.Deployment) bool {
		replicas := d.Spec.Replicas != nil && *d.Spec.Replicas == 0
		_, scaled := d.Annotations[_ReplicasAnnotation]
		return replicas && scaled
	})

	statefulSets, err := in.lister.StatefulSets.StatefulSets(namespace).List(labels.Everything())
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
			return in.updateStatefulSet(sts)
		}),
		util.ForEachE(deployments, func(_ int, deployment *apps.Deployment) error {
			replicas, _ := strconv.Atoi(deployment.ObjectMeta.Annotations[_ReplicasAnnotation])
			deployment.Spec.Replicas = util.Pointer(int32(replicas))
			delete(deployment.ObjectMeta.Annotations, _ReplicasAnnotation)
			return in.updateDeployment(deployment)
		}),
	)
}

func (in *Executor) updateDeployment(deployment *apps.Deployment) error {
	_, err := in.kube.CoreClient().
		AppsV1().
		Deployments(deployment.Namespace).
		Update(context.Background(), deployment, meta.UpdateOptions{})
	return err
}

func (in *Executor) updateStatefulSet(sts *apps.StatefulSet) error {
	_, err := in.kube.CoreClient().
		AppsV1().
		StatefulSets(sts.Namespace).
		Update(context.Background(), sts, meta.UpdateOptions{})
	return err
}

func (in *Executor) fetchNamespaces(filter string, reverse bool) ([]string, error) {
	list, err := in.lister.Namespaces.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	return SortNamespaces(list, filter, reverse), nil

}
