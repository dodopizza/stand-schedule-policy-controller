package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
