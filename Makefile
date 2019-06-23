.PHONY: bazel-build-images
bazel-build-images:
	CONTAINER_PREFIX=${CONTAINER_PREFIX} CONTAINER_TAG=${CONTAINER_TAG} ./hack/bazel/build-images.sh

.PHONY: bazel-push-images
bazel-push-images:
	CONTAINER_PREFIX=${CONTAINER_PREFIX} CONTAINER_TAG=${CONTAINER_TAG} ./hack/bazel/push-images.sh

.PHONY: deps-update
deps-update:
	./hack/deps-update.sh
