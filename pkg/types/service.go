/*
Copyright (C) GRyCAP - I3M - UPV

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

package types

import (
	"fmt"

	"github.com/goccy/go-yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// ContainerName name of the service container
	ContainerName = "oscar-container"

	// VolumeName name of the volume for mounting the OSCAR PVC
	VolumeName = "oscar-volume"

	// VolumePath path to mount the OSCAR PVC
	VolumePath = "/oscar/bin"

	// ConfigVolumeName name of the volume for mounting the service configMap
	ConfigVolumeName = "oscar-config"

	// ConfigPath path to mount the service configMap
	ConfigPath = "/oscar/config"

	// FDLFileName name of the FDL file to be stored in the service's configMap
	FDLFileName = "function_config.yaml"

	// ScriptFileName name of the user script file to be stored in the service's configMap
	ScriptFileName = "script.sh"

	// PVCName name of the OSCAR PVC
	PVCName = "oscar-pvc"

	// WatchdogName name of the OpenFaaS watchdog binary
	WatchdogName = "fwatchdog"

	// WatchdogProcess name of the environment variable used by the watchdog to handle requests
	WatchdogProcess = "fprocess"

	// SupervisorName name of the FaaS Supervisor binary
	SupervisorName = "supervisor"

	// ServiceLabel label for deploying services in all backs
	ServiceLabel = "oscar_service"

	// EventVariable name used by the environment variable where events are stored
	EventVariable = "EVENT"

	// OpenfaasZeroScalingLabel label to enable zero scaling in OpenFaaS functions
	OpenfaasZeroScalingLabel = "com.openfaas.scale.zero"
)

// Service represents an OSCAR service following the SCAR Function Definition Language
type Service struct {
	// The name of the service
	Name string `json:"name" binding:"required,max=39,min=1"`

	// Memory limit for the service following the kubernetes format
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-memory
	// Optional. (default: 256Mi)
	Memory string `json:"memory"`

	// CPU limit for the service following the kubernetes format
	// https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu
	// Optional. (default: 0.2)
	CPU string `json:"cpu"`

	// Log level for the FaaS Supervisor
	// Optional. (default: INFO)
	LogLevel string `json:"log_level"`

	// Docker image for the service
	Image string `json:"image" binding:"required"`

	// StorageIOConfig slices with the input and ouput service configuration
	// Optional
	Input  []StorageIOConfig `json:"input"`
	Output []StorageIOConfig `json:"output"`

	// The user script to execute when the service is invoked
	Script string `json:"script,omitempty" binding:"required"`

	// The user-defined environment variables assigned to the service
	// Optional
	Environment struct {
		Vars map[string]string `json:"Variables"`
	} `json:"environment"`

	// Configuration for the storage providers used by the service
	// Optional. (default: MinIOProvider["default"] with the server's config credentials)
	StorageProviders *StorageProviders `json:"storage_providers,omitempty"`
}

// ToPodSpec returns a k8s podSpec from the Service
func (service *Service) ToPodSpec() (*v1.PodSpec, error) {
	resources, err := createResources(service)
	if err != nil {
		return nil, err
	}

	podSpec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:  ContainerName,
				Image: service.Image,
				Env:   convertEnvVars(service.Environment.Vars),
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      VolumeName,
						ReadOnly:  true,
						MountPath: VolumePath,
					},
					{
						Name:      ConfigVolumeName,
						ReadOnly:  true,
						MountPath: ConfigPath,
					},
				},
				Command:   []string{"/bin/sh"},
				Args:      []string{"-c", fmt.Sprintf("%s/%s", VolumePath, WatchdogName)},
				Resources: resources,
			},
		},
		Volumes: []v1.Volume{
			{
				Name: VolumeName,
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: PVCName,
					},
				},
			},
			{
				Name: ConfigVolumeName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: service.Name,
						},
					},
				},
			},
		},
	}

	// Add the required environment variables for the watchdog
	addWatchdogEnvVars(podSpec)

	return podSpec, nil
}

// ToYAML returns the service as a Function Definition Language YAML
func (service Service) ToYAML() (string, error) {
	bytes, err := yaml.Marshal(service)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetMinIOWebhookARN returns the MinIO's notify_webhook ARN for the specified function
func (service *Service) GetMinIOWebhookARN() string {
	return fmt.Sprintf("arn:minio:sqs:%s:%s:webhook", service.StorageProviders.MinIO[DefaultProvider].Region, service.Name)
}

func convertEnvVars(vars map[string]string) []v1.EnvVar {
	envVars := []v1.EnvVar{}
	for k, v := range vars {
		envVars = append(envVars, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envVars
}

func createResources(service *Service) (v1.ResourceRequirements, error) {
	resources := v1.ResourceRequirements{
		Limits: v1.ResourceList{},
	}

	if len(service.CPU) > 0 {
		cpu, err := resource.ParseQuantity(service.CPU)
		if err != nil {
			return resources, err
		}
		resources.Limits[v1.ResourceCPU] = cpu
	}

	if len(service.Memory) > 0 {
		memory, err := resource.ParseQuantity(service.Memory)
		if err != nil {
			return resources, err
		}
		resources.Limits[v1.ResourceMemory] = memory
	}

	return resources, nil
}

func addWatchdogEnvVars(p *v1.PodSpec) {
	requiredEnvVars := []v1.EnvVar{
		// Use FaaS Supervisor to handle requests
		{
			Name:  WatchdogProcess,
			Value: fmt.Sprintf("%s/%s", VolumePath, SupervisorName),
		},
		// Other OpenFaaS Watchdog options
		// https://github.com/openfaas/faas/tree/master/watchdog
		// TODO: This should be configurable
		{
			Name:  "max_inflight",
			Value: "1",
		},
		{
			Name:  "write_debug",
			Value: "true",
		},
		{
			Name:  "exec_timeout",
			Value: "0",
		},
	}

	for i, cont := range p.Containers {
		if cont.Name == ContainerName {
			p.Containers[i].Env = append(p.Containers[i].Env, requiredEnvVars...)
		}
	}
}
