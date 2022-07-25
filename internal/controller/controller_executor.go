package controller

import (
	"context"
	"regexp"
	"time"

	"go.uber.org/zap"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

// todo: handle external resources
// todo: combine errors
// todo: reschedule after completion
// todo: check workitem.fireat and schedule.fireat

type (
	WorkItem struct {
		scheduleType string
		policyName   string
		fireAt       time.Time
	}
)

func (w *WorkItem) deadline() time.Time {
	timeout := time.Minute * 30
	return w.fireAt.Add(timeout)
}

func (c *Controller) execute(i interface{}) error {
	now := c.clock.Now()
	work := i.(WorkItem)

	if now.Before(work.fireAt) {
		c.logger.Warn("Skip execution of policy because of current time before scheduled",
			zap.String("policy_name", work.policyName),
			zap.String("work_type", work.scheduleType),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_at_time", work.fireAt))
		return nil
	}

	if now.After(work.deadline()) {
		c.logger.Warn("Skip execution of policy because of current time after deadline",
			zap.String("policy_name", work.policyName),
			zap.String("work_type", work.scheduleType),
			zap.Stringer("time", now),
			zap.Stringer("scheduled_deadline", work.deadline()))
		return nil
	}

	policy, err := c.lister.stands.Get(work.policyName)
	if errors.IsNotFound(err) {
		c.logger.Warn("Skip execution of policy because it not exists", zap.String("policy_name", work.policyName))
		return nil
	}

	if err != nil {
		return err
	}

	c.logger.Info("Run execution of policy",
		zap.String("policy_name", work.policyName),
		zap.String("work_type", work.scheduleType))

	switch work.scheduleType {
	case "startup":
		return c.executeStartup(policy)
	case "shutdown":
		return c.executeShutdown(policy)
	}

	c.logger.Warn("Invalid work type found for policy",
		zap.String("policy_name", work.policyName),
		zap.String("work_type", work.scheduleType))
	return nil
}

func (c *Controller) executeShutdown(policy *apis.StandSchedulePolicy) error {
	namespaces, err := c.filterNamespaces(policy.Spec.TargetNamespaceFilter)
	if err != nil {
		c.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	for _, namespace := range namespaces {
		quota := &core.ResourceQuota{
			ObjectMeta: meta.ObjectMeta{
				Name:      "zero-quota",
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

		c.logger.Debug("Create resource quota in namespace", zap.String("work_namespace", namespace))
		_, err := c.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Create(context.Background(), quota, meta.CreateOptions{})

		if errors.IsAlreadyExists(err) {
			continue
		}

		if err != nil {
			c.logger.Error("Failed to create resource quota in namespace",
				zap.Error(err),
				zap.String("work_namespace", namespace))
		}

		c.logger.Debug("Delete all existing pods in namespace", zap.String("work_namespace", namespace))
		err = c.kube.CoreClient().
			CoreV1().
			Pods(namespace).
			DeleteCollection(context.Background(), meta.DeleteOptions{}, meta.ListOptions{})

		if err != nil {
			c.logger.Error("Failed to delete all pods from namespace",
				zap.Error(err),
				zap.String("work_namespace", namespace))
		}
	}

	return nil
}

func (c *Controller) executeStartup(policy *apis.StandSchedulePolicy) error {
	namespaces, err := c.filterNamespaces(policy.Spec.TargetNamespaceFilter)
	if err != nil {
		c.logger.Warn("Failed to list target namespaces", zap.Error(err))
		return err
	}

	for _, namespace := range namespaces {
		err := c.kube.CoreClient().
			CoreV1().
			ResourceQuotas(namespace).
			Delete(context.Background(), "zero-quota", meta.DeleteOptions{})

		if errors.IsNotFound(err) {
			c.logger.Warn("Resource quota not exists in namespace", zap.String("work_namespace", namespace))
			continue
		}

		if err != nil {
			c.logger.Error("Failed to delete resource quota in namespace",
				zap.Error(err),
				zap.String("work_namespace", namespace))
		}
	}

	return nil
}

func (c *Controller) filterNamespaces(filter string) ([]string, error) {
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
