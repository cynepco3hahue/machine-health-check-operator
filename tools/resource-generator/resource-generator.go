/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"

	"github.com/openshift/machine-health-check-operator/tools/components"
	"github.com/openshift/machine-health-check-operator/tools/utils"
)

func main() {
	resourceType := flag.String("type", "", "Type of resource to generate. vmi | vmipreset | vmirs | vm | vmim | kv | rbac")
	namespace := flag.String("namespace", "kube-system", "Namespace to use.")
	repository := flag.String("repository", "kubevirt", "Image Repository to use.")
	version := flag.String("version", "latest", "version to use.")
	pullPolicy := flag.String("pullPolicy", "IfNotPresent", "ImagePullPolicy to use.")
	verbosity := flag.String("verbosity", "2", "Verbosity level to use.")

	flag.Parse()

	imagePullPolicy := v1.PullPolicy(*pullPolicy)

	switch *resourceType {
	case "machine-health-check-operator":
		operator, err := components.NewOperatorDeployment(*namespace, *repository, *version, imagePullPolicy, *verbosity)
		if err != nil {
			panic(fmt.Errorf("error generating machine-health-check-operator deployment %v", err))

		}
		utils.MarshallObject(operator, os.Stdout)
	default:
		panic(fmt.Errorf("unknown resource type %s", *resourceType))
	}
}
