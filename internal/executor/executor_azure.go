package executor

import (
	"context"
	"regexp"
	"sort"

	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

func (in *Executor) executeShutdownAzure(filters apis.AzureResourceList) error {
	sort.Sort(filters)

	resources, err := in.fetchAzureResources(context.Background(), filters)
	if err != nil {
		in.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	return util.ForEachE(util.MapKeys(resources), func(_ int, key int64) error {
		return util.ForEachParallelE(resources[key], func(_ int, resource *azure.Resource) error {
			return in.azure.Shutdown(context.Background(), resource, false)
		})
	})
}

func (in *Executor) executeStartupAzure(filters apis.AzureResourceList) error {
	sort.Sort(sort.Reverse(filters))

	resources, err := in.fetchAzureResources(context.Background(), filters)
	if err != nil {
		in.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	return util.ForEachE(util.MapKeys(resources), func(_ int, key int64) error {
		return util.ForEachParallelE(resources[key], func(_ int, resource *azure.Resource) error {
			return in.azure.Startup(context.Background(), resource, true)
		})
	})
}

func (in *Executor) fetchAzureResources(ctx context.Context, filters apis.AzureResourceList) (ret map[int64][]*azure.Resource, err error) {
	return ret, util.ForEachE(filters, func(_ int, filter apis.AzureResource) error {
		t, err := azure.From(filter.Type)
		if err != nil {
			return err
		}

		list, err := in.azure.List(ctx, t, filter.ResourceGroupName)
		if err != nil {
			return err
		}

		for _, resource := range list {
			match, _ := regexp.MatchString(filter.ResourceNameFilter, resource.GetName())

			if match {
				ret[filter.Priority] = append(ret[filter.Priority], resource)
			}
		}

		return nil
	})
}
