package pumba

import (
	"fmt"

	"github.com/faruqisan/resilia/pkg/kube"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CommandNetEm is worker command to run net em
	CommandNetEm WorkerCommandMode = "netem"

	// CommandNetEmDelay egress traffic
	CommandNetEmDelay NetEmCommands = "delay"
	// CommandNetEmLoss adds packet losses
	CommandNetEmLoss NetEmCommands = "loss"
	// CommandNetEmLossState adds packet losses, based on 4-state Markov probability model
	CommandNetEmLossState NetEmCommands = "loss-state"
	// CommandNetEmLossGEModel adds packet losses, according to the Gilbert-Elliot loss model
	CommandNetEmLossGEModel NetEmCommands = "loss-gemodel"
	// CommandNetEmRate limit egress traffic
	CommandNetEmRate NetEmCommands = "rate"
	// CommandNetEmDuplicate adds packet duplication
	CommandNetEmDuplicate NetEmCommands = "duplicate"
	// CommandNetEmCorrupt adds packet corruption
	CommandNetEmCorrupt NetEmCommands = "corrupt"

	// CommandPause is worker command to pause container
	CommandPause WorkerCommandMode = "pause"
)

type (
	// WorkerOptions function define worker options
	WorkerOptions func(worker *Worker)

	// WorkerCommandMode type define worker mode
	WorkerCommandMode string

	// Worker struct hold pumba worker data
	Worker struct {
		id       string
		target   string // target pod
		interval string
		mode     WorkerCommandMode

		// netem related
		netEmCommand NetEmCommands
		netEmOptions NetEmOptions

		// pause related
		pauseOptions PauseOptions
	}

	// NetEmCommands type define netem command
	NetEmCommands string
	// PauseCommands type define pause command
	PauseCommands string

	//NetEmOptions struct define netem options
	NetEmOptions struct {
		TCImage        string // Docker image with tc (iproute2 package); try 'gaiadocker/iproute2'
		Duration       string // network emulation duration; should be smaller than recurrent interval; use with optional unit suffix: 'ms/s/m/h'
		Interface      string // network interface to apply delay on (default: "eth0")
		TargetIPFilter string // target IP filter; supports multiple IPs; supports CIDR notation
		PullImage      bool   // try to pull tc-image

		// loss options
		LossPercent string
	}

	// PauseOptions struct define pause options
	PauseOptions struct {
		// Duration pause duration: must be shorter than recurrent interval;
		// use with optional unit suffix: 'ms/s/m/h'
		Duration string
	}

	// Engine struct act as function receiver and hold pumba engine
	// configuration
	Engine struct {
		kubeEngine *kube.Engine
	}
)

// New function return new engine struct with setuped configuration
func New(kubeEngine *kube.Engine) *Engine {
	return &Engine{
		kubeEngine: kubeEngine,
	}
}

// NetEm function set options to worker to use NetEm
func (e *Engine) NetEm(netEmCommand NetEmCommands, options NetEmOptions) WorkerOptions {
	return func(w *Worker) {
		w.netEmOptions = options
		w.mode = CommandNetEm
		w.netEmCommand = netEmCommand
	}
}

// Pause function set options to worker to use pause
func (e *Engine) Pause(options PauseOptions) WorkerOptions {
	return func(w *Worker) {
		w.pauseOptions = options
		w.mode = CommandPause
	}
}

// NewPumbaWorker function will spawn new pumba worker
func (e *Engine) NewPumbaWorker(target string, interval string, mode WorkerCommandMode, options ...WorkerOptions) Worker {
	var (
		w   Worker
		uID = uuid.New().ID()
	)

	w.target = target
	w.mode = mode
	w.interval = interval
	w.id = fmt.Sprintf("%d", uID)

	for _, option := range options {
		option(&w)
	}

	return w
}

// GetID function return worker id
func (w *Worker) GetID() string {
	return w.id
}

// RunWorker function will run given worker on k8s cluster
// and returning daemon name
func (e *Engine) RunWorker(worker Worker) (string, error) {
	ds := e.createPumbaDaemonSet(worker)
	return e.kubeEngine.CreateDaemonSet(ds)
}

// CreatePumbaDaemonSet function will create k8s daemonset object
// with given options
func (e *Engine) createPumbaDaemonSet(worker Worker) *appsv1.DaemonSet {
	var (
		d             *appsv1.DaemonSet
		daemonSetName = fmt.Sprintf("resilience-pumba-%s", worker.GetID())
		containerName = fmt.Sprintf("resilience-pumba-%s-container-%s", worker.mode, worker.GetID())
		volumeName    = "dockersocker"
		mountPath     = "/var/run/docker.sock"
		imageName     = "gaiaadm/pumba"
	)

	d = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: daemonSetName,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": daemonSetName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: daemonSetName,
					Labels: map[string]string{
						"app":               daemonSetName,
						"com.gaiaadm.pumba": "true",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  containerName,
							Image: imageName,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      volumeName,
									MountPath: mountPath,
								},
							},
							Args: worker.parseArgs(),
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name: volumeName,
							VolumeSource: apiv1.VolumeSource{
								HostPath: &apiv1.HostPathVolumeSource{
									Path: mountPath,
								},
							},
						},
					},
				},
			},
		},
	}

	return d
}

func (w Worker) parseArgs() []string {
	var (
		args       []string
		targetArgs = fmt.Sprintf("io.kubernetes.pod.name=%s", w.target)
	)

	args = []string{
		"--log-level", "info",
		"--label", targetArgs,
		"--interval", w.interval,
		string(w.mode),
	}

	switch w.mode {
	case CommandNetEm:
		args = append(args, w.parseNetEmArgs()...)
	case CommandPause:
		args = append(args, w.parsePauseArgs()...)
	}

	return args
}

func (w Worker) parseNetEmArgs() []string {
	args := []string{
		string(w.netEmCommand),
		"--duration", w.netEmOptions.Duration,
		"--tc-image", w.netEmOptions.TCImage,
	}

	if w.netEmOptions.Interface != "" {
		args = append(args, "-i", w.netEmOptions.Interface)
	}

	if w.netEmOptions.TargetIPFilter != "" {
		args = append(args, "-t", w.netEmOptions.TargetIPFilter)
	}

	if w.netEmOptions.LossPercent != "" {
		args = append(args, "--percent", w.netEmOptions.LossPercent)
	}

	return args
}

func (w Worker) parsePauseArgs() []string {
	var (
		args []string
	)

	if w.pauseOptions.Duration != "" {
		args = append(args, "-d", w.pauseOptions.Duration)
	}

	return args
}
