package executor

import (
	"context"
	"regexp"
	"strings"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	lcore "k8s.io/client-go/listers/core/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

// todo: handle external resources
// todo: scale deploy/sts before zero-quota

type (
	Executor struct {
		logger     *zap.Logger
		azure      azure.Interface
		kube       kubernetes.Interface
		namespaces lcore.NamespaceLister
	}
)

const (
	_ResourceQuotaName = "zero-quota"
)

func New(
	l *zap.Logger,
	az azure.Interface,
	k kubernetes.Interface,
	namespaces lcore.NamespaceLister,
) *Executor {
	return &Executor{
		logger:     l.Named("executor"),
		azure:      az,
		kube:       k,
		namespaces: namespaces,
	}
}

func (e *Executor) ExecuteShutdown(policy *apis.StandSchedulePolicy) error {
	namespaces, err := e.fetchNamespaces(policy.Spec.TargetNamespaceFilter, true)
	if err != nil {
		e.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	var summary error
	for _, namespace := range namespaces {
		e.logger.Debug("Create resource quota in namespace",
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
		_, err = e.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Create(context.Background(), quota, meta.CreateOptions{})
		summary = multierr.Append(summary, kubernetes.IgnoreAlreadyExistsError(err))

		e.logger.Debug("Delete all existing pods in namespace", zap.String("work_namespace", namespace))
		err = e.kube.CoreClient().
			CoreV1().
			Pods(namespace).
			DeleteCollection(context.Background(), meta.DeleteOptions{}, meta.ListOptions{})
		summary = multierr.Append(summary, err)
	}

	return summary
}

func (e *Executor) ExecuteStartup(policy *apis.StandSchedulePolicy) error {
	namespaces, err := e.fetchNamespaces(policy.Spec.TargetNamespaceFilter, false)
	if err != nil {
		e.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	var summary error
	for _, namespace := range namespaces {
		e.logger.Debug("Delete resource quota in namespace",
			zap.String("quota", _ResourceQuotaName),
			zap.String("namespace", namespace))

		err := e.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Delete(context.Background(), _ResourceQuotaName, meta.DeleteOptions{})
		summary = multierr.Append(summary, kubernetes.IgnoreNotFoundError(err))
	}

	return summary
}

func (e *Executor) fetchNamespaces(filter string, reverse bool) ([]string, error) {
	list, err := e.namespaces.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	return SortNamespaces(list, filter, reverse), nil

}

func SortNamespaces(objects []*core.Namespace, filter string, reverse bool) []string {
	var (
		namespaces []string
		filters    = strings.Split(filter, "|")
	)

	if reverse {
		filters = util.Reverse(filters)
	}

	for _, f := range filters {
		for _, namespace := range objects {
			matched, _ := regexp.MatchString(f, namespace.Name)

			if matched {
				namespaces = append(namespaces, namespace.Name)
			}
		}
	}

	return namespaces
}
