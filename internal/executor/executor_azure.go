package executor

import (
	"context"
	"sort"

	"go.uber.org/zap"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

func (ex *Executor) executeShutdownAzure(ctx context.Context, filters apis.AzureResourceList) error {
	resources, err := ex.fetchAzureResources(ctx, filters, false)
	if err != nil {
		ex.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	ex.logger.Debug("Shutdown azure resources")
	return util.ForEachE(filters, func(_ int, filter apis.AzureResource) error {
		return util.ForEachParallelE(resources[filter.Priority], func(_ int, resource *azure.Resource) error {
			ex.logger.Debug("Shutdown azure resource", zap.Stringer("resource", resource))
			return ex.azure.Shutdown(ctx, resource, false)
		})
	})
}

func (ex *Executor) executeStartupAzure(ctx context.Context, filters apis.AzureResourceList) error {
	resources, err := ex.fetchAzureResources(ctx, filters, true)
	if err != nil {
		ex.logger.Warn("Failed to list target azure resources", zap.Error(err))
		return err
	}

	ex.logger.Debug("Startup azure resources")
	return util.ForEachE(filters, func(_ int, filter apis.AzureResource) error {
		return util.ForEachParallelE(resources[filter.Priority], func(_ int, resource *azure.Resource) error {
			ex.logger.Debug("Startup azure resource", zap.Stringer("resource", resource))
			return ex.azure.Startup(ctx, resource, true)
		})
	})
}

func (ex *Executor) fetchAzureResources(ctx context.Context, filters apis.AzureResourceList, reverse bool) (map[int64][]*azure.Resource, error) {
	result := make(map[int64][]*azure.Resource)
	sortFilters := sort.Interface(filters)
	if reverse {
		sortFilters = sort.Reverse(sortFilters)
	}
	sort.Sort(sortFilters)

	return result, util.ForEachE(filters, func(_ int, filter apis.AzureResource) error {
		azureType, err := azure.From(filter.Type)
		if err != nil {
			return err
		}

		list, err := ex.azure.List(ctx, azureType, filter.ResourceGroupName)
		if err != nil {
			return err
		}

		MergeAzureResources(result, list, filter)
		return nil
	})
}
