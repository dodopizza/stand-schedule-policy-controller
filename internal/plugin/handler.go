package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/kubernetes"
	apis "github.com/dodopizza/stand-schedule-policy-controller/pkg/apis/standschedules/v1"
)

type (
	Handler struct {
		Type        apis.ConditionScheduleType
		Stand       string
		Wait        bool
		WaitTimeout time.Duration

		kube         kubernetes.Interface
		kubeFlags    *genericclioptions.ConfigFlags
		handlerFlags *pflag.FlagSet
	}
)

func NewStartupHandler() *Handler {
	return &Handler{
		Type:        apis.StatusStartup,
		WaitTimeout: time.Minute * 5,
	}
}

func NewShutdownHandler() *Handler {
	return &Handler{
		Type:        apis.StatusShutdown,
		WaitTimeout: time.Minute * 5,
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

func (h *Handler) Setup(stand string) error {
	k, err := kubernetes.NewPluginClient(h.kubeFlags)
	if err != nil {
		return err
	}
	h.kube = k
	h.Stand = stand
	return nil
}

func (h *Handler) Run() error {
	policy, err := h.fetchPolicy()
	if err != nil {
		return err
	}

	current := time.Now().UTC()
	override := current.Add(time.Second * 30).Round(time.Minute)
	fmt.Printf("Policy \"%s\" will be executed at: %s\n", policy.Name, override)

	schedule := policy.Spec.GetSchedule(h.Type)
	schedule.Override = override.Format(time.RFC3339)

	_, err = h.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		Update(context.Background(), policy, meta.UpdateOptions{})
	if err != nil {
		fmt.Printf("Failed to update policy \"%s\" definiton\n", policy.Name)
	}
	fmt.Printf("Policy \"%s\" definition updated\n", policy.Name)

	if h.Wait {
		fmt.Printf("Waiting to completion: ")
		return wait.PollImmediate(time.Second*15, h.WaitTimeout, h.WaitPolicyReady)
	}

	return nil
}

func (h *Handler) WaitPolicyReady() (bool, error) {
	fmt.Printf(".")

	policy, err := h.fetchPolicy()
	if err != nil {
		return false, err
	}

	status := policy.Status.GetScheduleStatus(h.Type)
	statusCompleted := status.Status == string(apis.ConditionCompleted) || status.Status == string(apis.ConditionFailed)

	if statusCompleted {
		fmt.Printf("\nDone\n")
	}

	return statusCompleted, nil
}

func (h *Handler) fetchPolicy() (*apis.StandSchedulePolicy, error) {
	return h.kube.StandSchedulesClient().
		StandSchedulesV1().
		StandSchedulePolicies().
		Get(context.Background(), h.Stand, meta.GetOptions{})
}
