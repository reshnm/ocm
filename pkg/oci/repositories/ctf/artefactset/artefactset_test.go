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

package artefactset_test

import (
	"archive/tar"
	"compress/gzip"
	"io"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/artefactset"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const MimeTypeOctetStream = "application/octet-stream"

func defaultManifestFill(a *artefactset.ArtefactSet) {
	art := artefactset.NewArtefact(a, nil)
	Expect(art.AddLayer(accessio.BlobAccessForData(MimeTypeOctetStream, []byte("testdata")), nil)).To(Equal(0))
	desc, err := art.Manifest()
	Expect(err).To(Succeed())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromString("testdata")))
	Expect(desc.Layers[0].MediaType).To(Equal(MimeTypeOctetStream))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := accessio.BlobAccessForData(MimeTypeOctetStream, []byte("{}"))
	Expect(a.AddBlob(config)).To(Succeed())
	desc.Config = *artdesc.DefaultBlobDescriptor(config)

	a.AddArtefact(art, nil)

}

var _ = Describe("artefact management", func() {
	var tempfs vfs.FileSystem
	var opts accessobj.Options

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t
		opts = accessobj.AccessOptions(accessobj.PathFileSystem(tempfs))
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("instantiate filesystem artefact", func() {
		a, err := artefactset.FormatDirectory.Create("test", opts, 0700)
		Expect(err).To(Succeed())
		Expect(vfs.DirExists(tempfs, "test/"+artefactset.BlobsDirectoryName)).To(BeTrue())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test/"+artefactset.ArtefactSetDescriptorFileName)).To(BeTrue())

		infos, err := vfs.ReadDir(tempfs, "test/"+artefactset.BlobsDirectoryName)
		Expect(err).To(Succeed())
		blobs := []string{}
		for _, fi := range infos {
			blobs = append(blobs, fi.Name())
		}
		Expect(blobs).To(ContainElements(
			"sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
	})

	It("instantiate tgz artefact", func() {
		a, err := artefactset.FormatTGZ.Create("test.tgz", opts, 0600)
		Expect(err).To(Succeed())

		defaultManifestFill(a)

		Expect(a.Close()).To(Succeed())
		Expect(vfs.FileExists(tempfs, "test.tgz")).To(BeTrue())

		file, err := tempfs.Open("test.tgz")
		Expect(err).To(Succeed())
		defer file.Close()
		zip, err := gzip.NewReader(file)
		Expect(err).To(Succeed())
		defer zip.Close()
		tr := tar.NewReader(zip)

		files := []string{}
		for {
			header, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				Fail(err.Error())
			}

			switch header.Typeflag {
			case tar.TypeDir:
				Expect(header.Name).To(Equal(artefactset.BlobsDirectoryName))
			case tar.TypeReg:
				files = append(files, header.Name)
			}
		}
		Expect(files).To(ContainElements(
			artefactset.ArtefactSetDescriptorFileName,
			"blobs/sha256.3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a",
			"blobs/sha256.44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			"blobs/sha256.810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"))
	})

	Context("manifest", func() {
		It("read from filesystem artefact", func() {
			a, err := artefactset.FormatDirectory.Create("test", opts, 0700)
			Expect(err).To(Succeed())
			Expect(vfs.DirExists(tempfs, "test/"+artefactset.BlobsDirectoryName)).To(BeTrue())
			defaultManifestFill(a)
			Expect(a.Close()).To(Succeed())

			a, err = artefactset.FormatDirectory.Open(accessobj.ACC_READONLY, "test", opts)
			defer a.Close()
			Expect(len(a.GetDescriptor().Manifests)).To(Equal(1))
			art, err := a.GetArtefact(a.GetDescriptor().Manifests[0].Digest)
			Expect(err).To(Succeed())
			Expect(art.IsManifest()).To(BeTrue())
			blob, err := art.GetBlob("sha256:810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50")
			Expect(err).To(Succeed())
			Expect(blob.Get()).To(Equal([]byte("testdata")))
			Expect(blob.MimeType()).To(Equal(MimeTypeOctetStream))
		})
	})
	Context("index", func() {

	})
})