package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
    "go_test",
)

go_library(
    name = "go_default_library",
    srcs = ["aws_credentials.go"],
    importpath = "k8s.io/kubernetes/pkg/credentialprovider/aws",
    deps = [
        "//pkg/credentialprovider:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/wait:go_default_library",
        "//staging/src/k8s.io/client-go/tools/cache:go_default_library",
        "//staging/src/k8s.io/component-base/version:go_default_library",
        "//staging/src/k8s.io/legacy-cloud-providers/aws:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/credentials:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/credentials/stscreds:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/ec2metadata:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/request:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/session:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/service/ecr:go_default_library",
        "//vendor/k8s.io/klog/v2:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/service/sts:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["aws_credentials_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//pkg/credentialprovider:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/service/ecr:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
