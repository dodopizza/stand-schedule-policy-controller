package executor

import (
	"regexp"
	"strings"

	core "k8s.io/api/core/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

func SortNamespaces(
	objects []*core.Namespace,
	filter string,
	reverse bool,
) []string {
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

func MergeAzureResources(
	result map[int64][]*azure.Resource,
	list []*azure.Resource,
	filter apis.AzureResource,
) {
	for _, resource := range list {
		match, _ := regexp.MatchString(filter.ResourceNameFilter, resource.GetName())

		if match {
			result[filter.Priority] = append(result[filter.Priority], resource)
		}
	}
}
