package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/azure"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
	"github.com/dodopizza/stand-schedule-policy-controller/pkg/util"
)

func Test_SortNamespaces(t *testing.T) {
	cases := []struct {
		name          string
		namespaces    []*core.Namespace
		filter        string
		reverse       bool
		expNamespaces []string
	}{
		{
			name: "direct order",
			namespaces: []*core.Namespace{
				{ObjectMeta: meta.ObjectMeta{Name: "ci"}},
				{ObjectMeta: meta.ObjectMeta{Name: "infra-some"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre-ru"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre-kz"}},
			},
			filter:        "^dev-sre$|dev-sre-[a-z]*",
			reverse:       false,
			expNamespaces: []string{"dev-sre", "dev-sre-ru", "dev-sre-kz"},
		},
		{
			name: "reverse order",
			namespaces: []*core.Namespace{
				{ObjectMeta: meta.ObjectMeta{Name: "ci"}},
				{ObjectMeta: meta.ObjectMeta{Name: "infra-some"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre-ru"}},
				{ObjectMeta: meta.ObjectMeta{Name: "dev-sre-kz"}},
			},
			filter:        "^dev-sre$|dev-sre-[a-z]*",
			reverse:       true,
			expNamespaces: []string{"dev-sre-ru", "dev-sre-kz", "dev-sre"},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			actual := SortNamespaces(tc.namespaces, tc.filter, tc.reverse)

			assert.Exactly(t, tc.expNamespaces, actual)
		})
	}
}

func Test_FilterAndMergeAzureResources(t *testing.T) {
	cases := []struct {
		name         string
		resources    []*azure.Resource
		filter       apis.AzureResource
		expResources []string
	}{
		{
			name: "filter all",
			resources: []*azure.Resource{
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-aa-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-bb-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-cc-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-dd-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-dd-ee-suffix"),
			},
			filter: apis.AzureResource{
				Type:               apis.AzureResourceManagedMySQL,
				ResourceGroupName:  "test",
				ResourceNameFilter: "test-mysql-.*-suffix",
				Priority:           1,
			},
			expResources: []string{
				"Microsoft.DBforMySQL/servers/test/test-mysql-aa-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-bb-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-cc-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-dd-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-dd-ee-suffix",
			},
		},
		{
			name: "exclude some",
			resources: []*azure.Resource{
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-ama-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-bob-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-clc-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-monolith-suffix"),
				azure.NewResource("/subscriptions/11111111-2222-3333-4444-555555555555/resourceGroups/test/providers/Microsoft.DBforMySQL/servers/test-mysql-monolith-ee-suffix"),
			},
			filter: apis.AzureResource{
				Type:               apis.AzureResourceManagedMySQL,
				ResourceGroupName:  "test",
				ResourceNameFilter: "test-mysql-(((?!monolith).)*)-suffix",
				Priority:           1,
			},
			expResources: []string{
				"Microsoft.DBforMySQL/servers/test/test-mysql-ama-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-bob-suffix",
				"Microsoft.DBforMySQL/servers/test/test-mysql-clc-suffix",
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			result := make(map[int64][]*azure.Resource)
			FilterAndMergeAzureResources(result, tc.resources, tc.filter)
			actual := util.Project(result[tc.filter.Priority],
				func(_ int, r *azure.Resource) string {
					return r.String()
				})

			assert.Exactly(t, tc.expResources, actual)
		})
	}
}
