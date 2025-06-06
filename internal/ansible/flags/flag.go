// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flags

import (
	"crypto/tls"
	"runtime"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Flags - Options to be used by an ansible operator
type Flags struct {
	ReconcilePeriod            time.Duration
	WatchesFile                string
	InjectOwnerRef             bool
	LeaderElection             bool
	MaxConcurrentReconciles    int
	AnsibleVerbosity           int
	AnsibleRolesPath           string
	AnsibleCollectionsPath     string
	MetricsBindAddress         string
	ProbeAddr                  string
	LeaderElectionResourceLock string
	LeaderElectionID           string
	LeaderElectionNamespace    string
	LeaseDuration              time.Duration
	RenewDeadline              time.Duration
	GracefulShutdownTimeout    time.Duration
	AnsibleArgs                string
	AnsibleLogEvents           string
	ProxyPort                  int
	EnableHTTP2                bool
	SecureMetrics              bool
	MetricsRequireRBAC         bool

	// If not nil, used to deduce which flags were set in the CLI.
	flagSet *pflag.FlagSet
}

const (
	AnsibleRolesPathEnvVar       = "ANSIBLE_ROLES_PATH"
	AnsibleCollectionsPathEnvVar = "ANSIBLE_COLLECTIONS_PATH"
)

// AddTo - Add the ansible operator flags to the the flagset
func (f *Flags) AddTo(flagSet *pflag.FlagSet) {
	// Store flagset internally to be used for lookups later.
	f.flagSet = flagSet

	// Ansible flags.
	flagSet.StringVar(&f.WatchesFile,
		"watches-file",
		"./watches.yaml",
		"Path to the watches file to use",
	)
	flagSet.BoolVar(&f.InjectOwnerRef,
		"inject-owner-ref",
		true,
		"The ansible operator will inject owner references unless this flag is false",
	)
	flagSet.IntVar(&f.AnsibleVerbosity,
		"ansible-verbosity",
		2,
		"Ansible verbosity. Overridden by environment variable.",
	)
	flagSet.StringVar(&f.AnsibleRolesPath,
		"ansible-roles-path",
		"",
		"Ansible Roles Path. If unset, roles are assumed to be in {{CWD}}/roles.",
	)
	flagSet.StringVar(&f.AnsibleCollectionsPath,
		"ansible-collections-path",
		"",
		"Path to installed Ansible Collections. If set, collections should be located in {{value}}/ansible_collections/. "+
			"If unset, collections are assumed to be in ~/.ansible/collections or /usr/share/ansible/collections.",
	)
	flagSet.StringVar(&f.AnsibleArgs,
		"ansible-args",
		"",
		"Ansible args. Allows user to specify arbitrary arguments for ansible-based operators.",
	)

	// Controller flags.
	flagSet.DurationVar(&f.ReconcilePeriod,
		"reconcile-period",
		10*time.Hour,
		"Default reconcile period for controllers",
	)
	flagSet.IntVar(&f.MaxConcurrentReconciles,
		"max-concurrent-reconciles",
		runtime.NumCPU(),
		"Maximum number of concurrent reconciles for controllers. Overridden by environment variable.",
	)

	// TODO(2.0.0): remove
	flagSet.StringVar(&f.MetricsBindAddress,
		"metrics-addr",
		":8443",
		"The address the metric endpoint binds to",
	)
	_ = flagSet.MarkDeprecated("metrics-addr", "use --metrics-bind-address instead")
	flagSet.StringVar(&f.MetricsBindAddress,
		"metrics-bind-address",
		":8443",
		"The address the metric endpoint binds to",
	)
	// TODO(2.0.0): for Go/Helm the port used is: 8081
	// update it to keep the project aligned to the other
	flagSet.StringVar(&f.ProbeAddr,
		"health-probe-bind-address",
		":6789",
		"The address the probe endpoint binds to.",
	)
	// TODO(2.0.0): remove
	flagSet.BoolVar(&f.LeaderElection,
		"enable-leader-election",
		false,
		"Enable leader election for controller manager. Enabling this will"+
			" ensure there is only one active controller manager.",
	)
	_ = flagSet.MarkDeprecated("enable-leader-election", "use --leader-elect instead")
	flagSet.BoolVar(&f.LeaderElection,
		"leader-elect",
		false,
		"Enable leader election for controller manager. Enabling this will"+
			" ensure there is only one active controller manager.",
	)
	flagSet.StringVar(&f.LeaderElectionID,
		"leader-election-id",
		"",
		"Name of the configmap that is used for holding the leader lock.",
	)
	flagSet.StringVar(&f.LeaderElectionNamespace,
		"leader-election-namespace",
		"",
		"Namespace in which to create the leader election configmap for"+
			" holding the leader lock (required if running locally with leader"+
			" election enabled).",
	)
	flagSet.StringVar(&f.LeaderElectionResourceLock,
		"leader-elect-resource-lock",
		"configmapsleases",
		"The type of resource object that is used for locking during leader election."+
			" Supported options are 'leases', 'endpointsleases' and 'configmapsleases'. Default is configmapsleases.",
	)
	_ = flagSet.MarkDeprecated("leader-elect-resource-lock", "This flag is now hardcoded to 'leases', which is the only supported option by client-go")
	flagSet.DurationVar(&f.LeaseDuration,
		"leader-elect-lease-duration",
		15*time.Second,
		"LeaseDuration is the duration that non-leader candidates will wait"+
			" to force acquire leadership. This is measured against time of last observed ack. Default is 15 seconds.",
	)
	flagSet.DurationVar(&f.RenewDeadline,
		"leader-elect-renew-deadline",
		10*time.Second,
		"RenewDeadline is the duration that the acting controlplane will retry"+
			" refreshing leadership before giving up. Default is 10 seconds.",
	)
	flagSet.DurationVar(&f.GracefulShutdownTimeout,
		"graceful-shutdown-timeout",
		30*time.Second,
		"The amount of time that will be spent waiting"+
			" for runners to gracefully exit.",
	)
	flagSet.StringVar(&f.AnsibleLogEvents,
		"ansible-log-events",
		"tasks",
		"Ansible log events. The log level for console logging."+
			" This flag can be set to either Nothing, Tasks, or Everything.",
	)
	flagSet.IntVar(&f.ProxyPort,
		"proxy-port",
		8888,
		"Ansible proxy server port. Defaults to 8888.",
	)
	flagSet.BoolVar(&f.EnableHTTP2,
		"enable-http2",
		false,
		"enables HTTP/2 on the webhook and metrics servers",
	)
	flagSet.BoolVar(&f.SecureMetrics,
		"metrics-secure",
		false,
		"enables secure serving of the metrics endpoint",
	)
	flagSet.BoolVar(&f.MetricsRequireRBAC,
		"metrics-require-rbac",
		false,
		"enables protection of the metrics endpoint with RBAC-based authn/authz."+
			"see https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/filters#WithAuthenticationAndAuthorization for more info")
}

// ToManagerOptions uses the flag set in f to configure options.
// Values of options take precedence over flag defaults,
// as values are assume to have been explicitly set.
func (f *Flags) ToManagerOptions(options manager.Options) manager.Options {
	// Alias FlagSet.Changed so options are still updated when fields are empty.
	changed := func(flagName string) bool {
		return f.flagSet.Changed(flagName)
	}
	if f.flagSet == nil {
		changed = func(flagName string) bool { return false }
	}

	// TODO(2.0.0): remove metrics-addr
	if changed("metrics-bind-address") || changed("metrics-addr") || options.Metrics.BindAddress == "" {
		options.Metrics.BindAddress = f.MetricsBindAddress
	}
	if changed("health-probe-bind-address") || options.HealthProbeBindAddress == "" {
		options.HealthProbeBindAddress = f.ProbeAddr
	}
	// TODO(2.0.0): remove enable-leader-election
	if changed("leader-elect") || changed("enable-leader-election") || !options.LeaderElection {
		options.LeaderElection = f.LeaderElection
	}
	if changed("leader-election-id") || options.LeaderElectionID == "" {
		options.LeaderElectionID = f.LeaderElectionID
	}
	if changed("leader-election-namespace") || options.LeaderElectionNamespace == "" {
		options.LeaderElectionNamespace = f.LeaderElectionNamespace
	}
	if changed("leader-elect-lease-duration") || options.LeaseDuration == nil {
		options.LeaseDuration = &f.LeaseDuration
	}
	if changed("leader-elect-renew-deadline") || options.RenewDeadline == nil {
		options.RenewDeadline = &f.RenewDeadline
	}
	if options.LeaderElectionResourceLock == "" {
		options.LeaderElectionResourceLock = resourcelock.LeasesResourceLock
	}
	if changed("graceful-shutdown-timeout") || options.GracefulShutdownTimeout == nil {
		options.GracefulShutdownTimeout = &f.GracefulShutdownTimeout
	}

	disableHTTP2 := func(c *tls.Config) {
		c.NextProtos = []string{"http/1.1"}
	}
	if !f.EnableHTTP2 {
		options.WebhookServer = webhook.NewServer(webhook.Options{
			TLSOpts: []func(*tls.Config){disableHTTP2},
		})
		options.Metrics.TLSOpts = append(options.Metrics.TLSOpts, disableHTTP2)
	}
	options.Metrics.SecureServing = f.SecureMetrics

	if f.MetricsRequireRBAC {
		// FilterProvider is used to protect the metrics endpoint with authn/authz.
		// These configurations ensure that only authorized users and service accounts
		// can access the metrics endpoint. The RBAC are configured in 'config/rbac/kustomization.yaml'. More info:
		// https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/metrics/filters#WithAuthenticationAndAuthorization
		options.Metrics.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	return options
}
