package executor

import (
	"strings"

	"github.com/dlclark/regexp2"

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
		reg, err := regexp2.Compile(f, regexp2.None)

		if err != nil {
			continue
		}

		for _, namespace := range objects {
			matched, _ := reg.MatchString(namespace.Name)

			if matched {
				namespaces = append(namespaces, namespace.Name)
			}
		}
	}

	return namespaces
}

func FilterAndMergeAzureResources(
	result map[int64][]*azure.Resource,
	list []*azure.Resource,
	filter apis.AzureResource,
) {
	reg, err := regexp2.Compile(filter.ResourceNameFilter, regexp2.None)

	if err != nil {
		return
	}

	for _, resource := range list {
		match, _ := reg.MatchString(resource.GetName())

		if match {
			result[filter.Priority] = append(result[filter.Priority], resource)
		}
	}
}
