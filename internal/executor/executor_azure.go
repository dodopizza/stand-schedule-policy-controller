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

func (ex *Executor) executeShutdownAzure(filters apis.AzureResourceList) error {
	sort.Sort(filters)

	resources, err := ex.fetchAzureResources(context.Background(), filters)
	if err != nil {
		ex.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	return util.ForEachE(util.MapKeys(resources), func(_ int, key int64) error {
		return util.ForEachParallelE(resources[key], func(_ int, resource *azure.Resource) error {
			return ex.azure.Shutdown(context.Background(), resource, false)
		})
	})
}

func (ex *Executor) executeStartupAzure(filters apis.AzureResourceList) error {
	sort.Sort(sort.Reverse(filters))

	resources, err := ex.fetchAzureResources(context.Background(), filters)
	if err != nil {
		ex.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	return util.ForEachE(util.MapKeys(resources), func(_ int, key int64) error {
		return util.ForEachParallelE(resources[key], func(_ int, resource *azure.Resource) error {
			return ex.azure.Startup(context.Background(), resource, true)
		})
	})
}

func (ex *Executor) fetchAzureResources(ctx context.Context, filters apis.AzureResourceList) (map[int64][]*azure.Resource, error) {
	ret := make(map[int64][]*azure.Resource)

	return ret, util.ForEachE(filters, func(_ int, filter apis.AzureResource) error {
		azureType, err := azure.From(filter.Type)
		if err != nil {
			return err
		}

		list, err := ex.azure.List(ctx, azureType, filter.ResourceGroupName)
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
