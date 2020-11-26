package options

import (
	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

var (
	ProxyCIDRWhitelist kubeoptions.IPNetSlice = kubeoptions.DefaultProxyCIDRWhitelist
)
