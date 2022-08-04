package executor

import (
	"regexp"
	"strings"

	core "k8s.io/api/core/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

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
