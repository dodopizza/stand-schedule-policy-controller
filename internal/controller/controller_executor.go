package controller

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

// todo: handle external resources
// todo: check workitem.fireat and schedule.fireat

type (
	WorkItem struct {
		policyName   string
		scheduleType apis.ConditionScheduleType
		fireAt       time.Time
	}
)

const (
	_ResourceQuotaName = "zero-quota"
)

func (w *WorkItem) deadline() time.Time {
	timeout := time.Minute * 30
	return w.fireAt.Add(timeout)
}

func (c *Controller) execute(i interface{}) error {
	now := c.clock.Now()
	item := i.(WorkItem)

	if now.Before(item.fireAt) {
		c.logger.Warn("Skip execution of policy because of current time before scheduled",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_at_time", item.fireAt))
		return nil
	}

	if now.After(item.deadline()) {
		c.logger.Warn("Skip execution of policy because of current time after deadline",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_deadline", item.deadline()))
		return nil
	}

	state, exists := c.state.Get(item.policyName)
	policy, err := c.lister.stands.Get(item.policyName)
	if errors.IsNotFound(err) || !exists {
		c.logger.Warn("Skip execution of policy because it not exists", zap.String("policy_name", item.policyName))
		return nil
	}

	if err != nil {
		return err
	}

	c.logger.Info("Execute schedule of policy",
		zap.String("policy_name", item.policyName),
		zap.String("schedule_type", string(item.scheduleType)))

	switch item.scheduleType {
	case apis.StatusShutdown:
		err = c.executeShutdown(policy)
		state.UpdateStatus(now, err, item.scheduleType)
	case apis.StatusStartup:
		err = c.executeStartup(policy)
		state.UpdateStatus(now, err, item.scheduleType)
	default:
		err = fmt.Errorf("not supported schedule type specified: %s", item.scheduleType)
	}

	if err != nil {
		c.logger.Error("Failed to execute schedule of policy",
			zap.String("policy_name", item.policyName),
			zap.String("schedule_type", string(item.scheduleType)),
			zap.Error(err))
		return err
	}

	return nil
}

func (c *Controller) executeShutdown(policy *apis.StandSchedulePolicy) error {
	namespaces, err := c.fetchNamespaces(policy.Spec.TargetNamespaceFilter)
	if err != nil {
		c.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	var summary error
	for _, namespace := range namespaces {
		c.logger.Debug("Create resource quota in namespace",
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
		quota, err = c.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Create(context.Background(), quota, meta.CreateOptions{})
		summary = multierr.Append(summary, kubernetes.IgnoreAlreadyExistsError(err))

		c.logger.Debug("Delete all existing pods in namespace", zap.String("work_namespace", namespace))
		err = c.kube.CoreClient().
			CoreV1().
			Pods(namespace).
			DeleteCollection(context.Background(), meta.DeleteOptions{}, meta.ListOptions{})
		summary = multierr.Append(summary, err)
	}

	return summary
}

func (c *Controller) executeStartup(policy *apis.StandSchedulePolicy) error {
	namespaces, err := c.fetchNamespaces(policy.Spec.TargetNamespaceFilter)
	if err != nil {
		c.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	var summary error
	for _, namespace := range namespaces {
		c.logger.Debug("Delete resource quota in namespace",
			zap.String("quota", _ResourceQuotaName),
			zap.String("namespace", namespace))

		err := c.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Delete(context.Background(), _ResourceQuotaName, meta.DeleteOptions{})
		summary = multierr.Append(summary, kubernetes.IgnoreNotFoundError(err))
	}

	return summary
}

func (c *Controller) fetchNamespaces(filter string) ([]string, error) {
	result := []string{}

	namespaces, err := c.lister.ns.List(labels.Everything())
	if err != nil {
		return result, err
	}

	for _, namespace := range namespaces {
		matched, err := regexp.MatchString(filter, namespace.Name)

		if err != nil {
			continue
		}

		if matched {
			result = append(result, namespace.Name)
		}
	}

	return result, nil
}
