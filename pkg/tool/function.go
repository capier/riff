/*
 * Copyright 2018 The original author or authors
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
 */

package tool

import (
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateFunctionOptions struct {
	Name      string
	Namespace string
	Image     string
}

func (c *client) CreateFunction(options CreateFunctionOptions) (*v1alpha1.Service, error) {
	s := v1alpha1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      options.Name,
			Namespace: options.Namespace,
		},
		Spec: v1alpha1.ServiceSpec{
			RunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					RevisionTemplate: v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							Container: core_v1.Container{
								Image: options.Image,
							},
						},
					},
				},
			},
		},
	}

	svc, err := c.serving.ServingV1alpha1().Services(options.Namespace).Create(&s)

	return svc, err
}
