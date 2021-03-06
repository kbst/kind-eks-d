/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"errors"
	"fmt"
	"net"
	"strings"

	apiextensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	genericfeatures "k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/features"
	netutils "k8s.io/utils/net"
)

func validateProxyCIDRAllowlist(options *ServerRunOptions) []error {
	errors := []error{}

	// if its empty, don't add any IPs to the list
	for _, cidr := range options.ProxyCIDRAllowlist {
		if cidr.IP == nil {
			errors = append(errors, fmt.Errorf("invalid --proxy-cidr-allowlist (or --proxy-cidr-whitelist) specified"))
		}
	}
	return errors
}

// TODO: Longer term we should read this from some config store, rather than a flag.
// validateClusterIPFlags is expected to be called after Complete()
func validateClusterIPFlags(options *ServerRunOptions) []error {
	var errs []error
	// maxCIDRBits is used to define the maximum CIDR size for the cluster ip(s)
	const maxCIDRBits = 20

	// validate that primary has been processed by user provided values or it has been defaulted
	if options.PrimaryServiceClusterIPRange.IP == nil {
		errs = append(errs, errors.New("--service-cluster-ip-range must contain at least one valid cidr"))
	}

	serviceClusterIPRangeList := strings.Split(options.ServiceClusterIPRanges, ",")
	if len(serviceClusterIPRangeList) > 2 {
		errs = append(errs, errors.New("--service-cluster-ip-range must not contain more than two entries"))
	}

	// Complete() expected to have set Primary* and Secondary*
	// primary CIDR validation
	if err := validateMaxCIDRRange(options.PrimaryServiceClusterIPRange, maxCIDRBits, "--service-cluster-ip-range"); err != nil {
		errs = append(errs, err)
	}

	// Secondary IP validation
	secondaryServiceClusterIPRangeUsed := (options.SecondaryServiceClusterIPRange.IP != nil)
	if secondaryServiceClusterIPRangeUsed && !utilfeature.DefaultFeatureGate.Enabled(features.IPv6DualStack) {
		errs = append(errs, fmt.Errorf("--secondary-service-cluster-ip-range can only be used if %v feature is enabled", string(features.IPv6DualStack)))
	}

	// note: While the cluster might be dualstack (i.e. pods with multiple IPs), the user may choose
	// to only ingress traffic within and into the cluster on one IP family only. this family is decided
	// by the range set on --service-cluster-ip-range. If/when the user decides to use dual stack services
	// the Secondary* must be of different IPFamily than --service-cluster-ip-range
	if secondaryServiceClusterIPRangeUsed {
		// Should be dualstack IPFamily(PrimaryServiceClusterIPRange) != IPFamily(SecondaryServiceClusterIPRange)
		dualstack, err := netutils.IsDualStackCIDRs([]*net.IPNet{&options.PrimaryServiceClusterIPRange, &options.SecondaryServiceClusterIPRange})
		if err != nil {
			errs = append(errs, errors.New("error attempting to validate dualstack for --service-cluster-ip-range and --secondary-service-cluster-ip-range"))
		}

		if !dualstack {
			errs = append(errs, errors.New("--service-cluster-ip-range and --secondary-service-cluster-ip-range must be of different IP family"))
		}

		if err := validateMaxCIDRRange(options.SecondaryServiceClusterIPRange, maxCIDRBits, "--secondary-service-cluster-ip-range"); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func validateMaxCIDRRange(cidr net.IPNet, maxCIDRBits int, cidrFlag string) error {
	// Should be smallish sized cidr, this thing is kept in etcd
	// bigger cidr (specially those offered by IPv6) will add no value
	// significantly increase snapshotting time.
	var ones, bits = cidr.Mask.Size()
	if bits-ones > maxCIDRBits {
		return fmt.Errorf("specified %s is too large; for %d-bit addresses, the mask must be >= %d", cidrFlag, bits, bits-maxCIDRBits)
	}

	return nil
}

func validateServiceNodePort(options *ServerRunOptions) []error {
	var errs []error

	if options.KubernetesServiceNodePort < 0 || options.KubernetesServiceNodePort > 65535 {
		errs = append(errs, fmt.Errorf("--kubernetes-service-node-port %v must be between 0 and 65535, inclusive. If 0, the Kubernetes master service will be of type ClusterIP", options.KubernetesServiceNodePort))
	}

	if options.KubernetesServiceNodePort > 0 && !options.ServiceNodePortRange.Contains(options.KubernetesServiceNodePort) {
		errs = append(errs, fmt.Errorf("kubernetes service port range %v doesn't contain %v", options.ServiceNodePortRange, (options.KubernetesServiceNodePort)))
	}
	return errs
}

func validateTokenRequest(options *ServerRunOptions) []error {
	var errs []error

	enableAttempted := options.ServiceAccountSigningKeyFile != "" ||
		options.Authentication.ServiceAccounts.Issuer != "" ||
		len(options.Authentication.APIAudiences) != 0

	enableSucceeded := options.ServiceAccountIssuer != nil

	if enableAttempted && !utilfeature.DefaultFeatureGate.Enabled(features.TokenRequest) {
		errs = append(errs, errors.New("the TokenRequest feature is not enabled but --service-account-signing-key-file, --service-account-issuer and/or --api-audiences flags were passed"))
	}

	if utilfeature.DefaultFeatureGate.Enabled(features.BoundServiceAccountTokenVolume) && !utilfeature.DefaultFeatureGate.Enabled(features.TokenRequest) {
		errs = append(errs, errors.New("the BoundServiceAccountTokenVolume feature depends on the TokenRequest feature, but the TokenRequest features is not enabled"))
	}

	if !enableAttempted && utilfeature.DefaultFeatureGate.Enabled(features.BoundServiceAccountTokenVolume) {
		errs = append(errs, errors.New("--service-account-signing-key-file and --service-account-issuer are required flags"))
	}

	if enableAttempted && !enableSucceeded {
		errs = append(errs, errors.New("--service-account-signing-key-file, --service-account-issuer, and --api-audiences should be specified together"))
	}

	return errs
}

func validateAPIPriorityAndFairness(options *ServerRunOptions) []error {
	if utilfeature.DefaultFeatureGate.Enabled(genericfeatures.APIPriorityAndFairness) && options.GenericServerRunOptions.EnablePriorityAndFairness {
		// We need the alpha API enabled.  There are only a few ways to turn it on
		enabledAPIString := options.APIEnablement.RuntimeConfig.String()
		switch {
		case strings.Contains(enabledAPIString, "api/all=true"):
			return nil
		case strings.Contains(enabledAPIString, "api/alpha=true"):
			return nil
		case strings.Contains(enabledAPIString, "flowcontrol.apiserver.k8s.io/v1alpha1=true"):
			return nil
		default:
			return []error{fmt.Errorf("enabling APIPriorityAndFairness requires --runtime-confg=flowcontrol.apiserver.k8s.io/v1alpha1=true to enable the required API")}
		}
	}

	return nil
}

// Validate checks ServerRunOptions and return a slice of found errs.
func (s *ServerRunOptions) Validate() []error {
	var errs []error
	if s.MasterCount <= 0 {
		errs = append(errs, fmt.Errorf("--apiserver-count should be a positive number, but value '%d' provided", s.MasterCount))
	}
	errs = append(errs, s.Etcd.Validate()...)
	if es := validateProxyCIDRAllowlist(s); len(es) > 0 {
		errs = append(errs, es...)
	}
	errs = append(errs, validateClusterIPFlags(s)...)
	errs = append(errs, validateServiceNodePort(s)...)
	errs = append(errs, validateAPIPriorityAndFairness(s)...)
	errs = append(errs, s.SecureServing.Validate()...)
	errs = append(errs, s.Authentication.Validate()...)
	errs = append(errs, s.Authorization.Validate()...)
	errs = append(errs, s.Audit.Validate()...)
	errs = append(errs, s.Admission.Validate()...)
	errs = append(errs, s.InsecureServing.Validate()...)
	errs = append(errs, s.APIEnablement.Validate(legacyscheme.Scheme, apiextensionsapiserver.Scheme, aggregatorscheme.Scheme)...)
	errs = append(errs, validateTokenRequest(s)...)
	errs = append(errs, s.Metrics.Validate()...)
	errs = append(errs, s.Logs.Validate()...)

	return errs
}
