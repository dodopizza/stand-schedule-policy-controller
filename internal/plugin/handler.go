package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	Handler struct {
		Type apis.ConditionScheduleType
		Wait bool

		kube         kubernetes.Interface
		kubeFlags    *genericclioptions.ConfigFlags
		handlerFlags *pflag.FlagSet
	}
)

func NewStartupHandler() *Handler {
	return &Handler{
		Type: apis.StatusStartup,
	}
}

func NewShutdownHandler() *Handler {
	return &Handler{
		Type: apis.StatusShutdown,
	}
}

func (h *Handler) String() string {
	return strings.ToLower(string(h.Type))
}

func (h *Handler) SetupFlags() *pflag.FlagSet {
	h.handlerFlags = pflag.NewFlagSet(h.String(), pflag.ExitOnError)
	h.handlerFlags.BoolVar(&h.Wait, "wait", h.Wait, "Wait to completion")
	h.kubeFlags = genericclioptions.NewConfigFlags(false)
	h.kubeFlags.AddFlags(h.handlerFlags)

	return h.handlerFlags
}

func (h *Handler) Setup() error {
	k, err := kubernetes.NewPluginClient(h.kubeFlags)
	if err != nil {
		return err
	}
	h.kube = k
	return nil
}

func (h *Handler) Run(stand string) error {
	policy, err := h.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		Get(context.Background(), stand, meta.GetOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("Found sspol %s\n", policy.Name)

	current := time.Now().UTC()
	override := current.Add(time.Second * 30).Round(time.Minute)
	fmt.Printf("Generated override time for policy %s is %s\n", policy.Name, override)

	switch h.Type {
	case apis.StatusStartup:
		policy.Spec.Schedules.Startup.Override = override.Format(time.RFC3339)
	case apis.StatusShutdown:
		policy.Spec.Schedules.Shutdown.Override = override.Format(time.RFC3339)
	default:
		return fmt.Errorf("invalid type %s specified\n", h.Type)
	}

	fmt.Printf("Update sspol %s definition\n", policy.Name)
	_, err = h.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		Update(context.Background(), policy, meta.UpdateOptions{})
	if err != nil {
		fmt.Printf("Failed to update sspol %s definiton\n", policy.Name)
	}

	if h.Wait {
		// todo: wait support
	}

	return nil
}
