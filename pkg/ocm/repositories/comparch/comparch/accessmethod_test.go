// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package comparch_test

import (
	"encoding/json"
	"os"
	"reflect"

	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/accessmethods"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	comparch2 "github.com/gardener/ocm/pkg/ocm/repositories/comparch/comparch"
	"github.com/gardener/ocm/pkg/runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var DefaultContext = ocm.New()

var _ = Describe("access method", func() {

	legacy := "{\"type\":\"localFilesystemBlob\",\"fileName\":\"anydigest\",\"mediaType\":\"application/json\"}"

	Context("local access method", func() {
		It("decodes legacy methood", func() {
			spec, err := DefaultContext.AccessSpecForConfig([]byte(legacy), nil)
			Expect(err).To(Succeed())
			Expect(reflect.TypeOf(spec).String()).To(Equal("*accessmethods.LocalBlobAccessSpec"))
			Expect(spec.(*accessmethods.LocalBlobAccessSpec).ReferenceName).To(Equal("anydigest"))
		})

		It("encodes legacy methood", func() {
			spec := comparch2.NewLocalFilesystemBlobAccessSpecV1("anydigest", "application/json")
			data, err := DefaultContext.Encode(spec, runtime.DefaultJSONEncoding)
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte(legacy)))
		})
	})

	Context("component archive", func() {
		It("instantiate local blob access method for component archive", func() {
			data, err := os.ReadFile("testdata/component-descriptor.yaml")
			Expect(err).To(Succeed())
			cd, err := compdesc.Decode(data)
			Expect(err).To(Succeed())

			ca, err := comparch2.New(DefaultContext, accessobj.ACC_CREATE, nil, nil, nil, 0600)
			Expect(err).To(Succeed())

			ca.GetDescriptor().Name = "acme.org/dummy"
			ca.GetDescriptor().Version = "v1"

			res, err := cd.GetResourceByIdentity(metav1.NewIdentity("local"))
			Expect(err).To(Succeed())
			Expect(res).To(Not(BeNil()))

			spec, err := DefaultContext.AccessSpecForSpec(res.Access)
			Expect(err).To(Succeed())
			Expect(spec).To(Not(BeNil()))

			Expect(spec.GetType()).To(Equal(comparch2.LocalFilesystemBlobType))
			Expect(spec.GetKind()).To(Equal(comparch2.LocalFilesystemBlobType))
			Expect(spec.GetVersion()).To(Equal("v1"))
			Expect(reflect.TypeOf(spec).String()).To(Equal("*accessmethods.LocalBlobAccessSpec"))

			data, err = json.Marshal(spec)
			Expect(err).To(Succeed())
			Expect(string(data)).To(Equal(legacy))

			m, err := spec.AccessMethod(ca)
			Expect(err).To(Succeed())
			Expect(m).To(Not(BeNil()))
			Expect(reflect.TypeOf(m).String()).To(Equal("*comparch.localFilesystemBlobAccessMethod"))
			Expect(m.GetKind()).To(Equal("localBlob"))

			Expect(ca.Close()).To(Succeed())
		})
	})
})