.PHONY: bazel-build-images
bazel-build:
	./hack/dockerized "./hack/bazel/build.sh"

.PHONY: bazel-push-images
bazel-push-images:
	./hack/dockerized "CONTAINER_PREFIX=${CONTAINER_PREFIX} CONTAINER_TAG=${CONTAINER_TAG} ./hack/bazel/push-images.sh"

.PHONY: fmt
fmt:
	./hack/dockerized "./hack/bazel/fmt.sh"

.PHONY: deps-update
deps-update:
	./hack/dockerized "SYNC_VENDOR=true ./hack/deps-update.sh"
