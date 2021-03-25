package options

import (
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

var (
	ProxyCIDRAllowlist kubeoptions.IPNetSlice = kubeoptions.DefaultProxyCIDRAllowlist
)
