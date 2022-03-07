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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/containerd/containerd/reference/docker"
	"github.com/containers/image/v5/docker/archive"
	"github.com/containers/image/v5/docker/daemon"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/config"
	"github.com/gardener/ocm/pkg/credentials"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/ociutils"
	docker2 "github.com/gardener/ocm/pkg/oci/repositories/docker"
	_ "github.com/gardener/ocm/pkg/ocm"
)

const MIME_OCTET = "application/octet-stream"

func Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+msg, args...)
	os.Exit(1)
}

func handleError(err error, msg string, args ...interface{}) {
	if err != nil {
		Error("%s; %s", fmt.Sprintf(msg, args...), err)
	}
}

func setupCredentials() {
	data, err := ioutil.ReadFile("config.yaml")
	handleError(err, "cannot read config")

	cfg, err := config.DefaultContext().GetConfigForData(data, nil)
	handleError(err, "invalid config")
	err = config.DefaultContext().ApplyConfig(cfg, "config.yaml")
	handleError(err, "invalid config")
	ctx := credentials.DefaultContext()

	/*

		spec := dockerconfig.NewRepositorySpec(os.Getenv("HOME") + "/.docker/config.json").WithConsumerPropagation(true)
		repo, err := ctx.RepositoryForSpec(spec)
		handleError(err, "setup credentials")
		_ = repo
	*/
	_ = ctx
}

func dockerwritetest() {

	os.Remove("/tmp/docker.tar")
	w, err := archive.NewWriter(nil, "/tmp/docker.tar")
	handleError(err, "writer")

	defer w.Close()

	any, err := docker.ParseAnyReference("ghcr.io/mandelsoft/pause:test")
	handleError(err, "ref")

	ref, err := w.NewReference(any.(reference.NamedTagged))
	handleError(err, "attach writer")

	dst, err := ref.NewImageDestination(context.Background(), nil)
	handleError(err, "dest")

	ctx := oci.DefaultContext()

	version := "0.1-dev"
	spec := docker2.NewRepositorySpec()
	name := "ghcr.io/mandelsoft/pause"

	repo, err := ctx.RepositoryForSpec(spec)
	handleError(err, "get repo")

	ns, err := repo.LookupNamespace(name)
	handleError(err, "lookup namespace")

	defer ns.Close()

	art, err := ns.GetArtefact(version)
	handleError(err, "lookup artefact")

	defer art.Close()

	err = docker2.Convert(art, nil, dst)
	handleError(err, "convert")
	err = dst.Commit(context.Background(), nil)
	handleError(err, "commit")

	/*
		m, mime, err := img.Manifest(context.Background())
		handleError(err, "manifest")
		fmt.Printf("manifest [%s]: %s\n", mime, string(m))

		cfg, err := img.ConfigBlob(context.Background())
		handleError(err, "config")
		fmt.Printf("config: %s\n", string(cfg))
	*/

}

func dockertest() {

	ref, err := archive.ParseReference("ghcr.io/mandelsoft/pause:0.1-dev")
	ref, err = daemon.ParseReference("ghcr.io/mandelsoft/pause:0.1-dev")
	ref, err = daemon.NewReference("c4c442d0040d", nil)
	ref, err = daemon.NewReference("ca617b241345", nil)
	handleError(err, "ref")

	//ref.NewImageDestination(context.Background(), nil)
	src, err := ref.NewImageSource(context.Background(), nil)
	handleError(err, "source")

	defer src.Close()

	data, mime, err := src.GetManifest(context.Background(), nil)
	handleError(err, "manifest")

	fmt.Printf("mime: %s\n", mime)
	fmt.Printf("manifest:\n %s\n*********\n", string(data))

	opts := types.ManifestUpdateOptions{
		ManifestMIMEType: artdesc.MediaTypeImageManifest,
	}
	un := image.UnparsedInstance(src, nil)
	img, err := image.FromUnparsedImage(context.Background(), nil, un)
	handleError(err, "manifest")

	img, err = img.UpdatedImage(context.Background(), opts)
	handleError(err, "convert")

	data, mime, err = img.Manifest(context.Background())
	handleError(err, "manifest")

	fmt.Printf("mime: %s\n", mime)
	fmt.Printf("manifest:\n %s\n*********\n", string(data))

	art, err := artdesc.Decode(data)
	handleError(err, "decode")

	for i, l := range art.Manifest().Layers {
		fmt.Printf("  layer %d [%s]: %s\n", i, l.MediaType, l.Digest)
	}
	info := img.LayerInfos()
	handleError(err, "layer info")
	for i, l := range info {
		fmt.Printf("  layer %d [%s]: %s\n", i, l.MediaType, l.Digest)
	}
	os.Exit(0)
}

func Print(resourceURl string) {
	fmt.Printf("%s:\n", resourceURl)
	/*
		ref, err := docker.ParseDockerRef(resourceURl)
		if err == nil {
			fmt.Printf("  name:   %s\n", ref.Name())
			fmt.Printf("  domain: %s\n", docker.Domain(ref) )
			fmt.Printf("  path:   %s\n", docker.Path(ref) )

			if t, ok := ref.(docker.Tagged); ok {
				fmt.Printf("  tag:    %s\n", t.Tag() )
			}
			if t, ok := ref.(docker.Digested); ok {
				fmt.Printf("  digest:  %s\n", t.Digest() )
			}
		} else {
			fmt.Printf("  err:    %s\n", err)
		}
	*/
	a, err := docker.ParseAnyReference(resourceURl)
	if err == nil {
		fmt.Printf("  any:   %s\n", a.String())
		if t, ok := a.(docker.Named); ok {
			fmt.Printf("  name:   %s\n", t.Name())
			fmt.Printf("  domain: %s\n", docker.Domain(t))
			fmt.Printf("  path:   %s\n", docker.Path(t))
		}
		if t, ok := a.(docker.Tagged); ok {
			fmt.Printf("  tag:    %s\n", t.Tag())
		}
		if t, ok := a.(docker.Digested); ok {
			fmt.Printf("  digest:  %s\n", t.Digest())
		}
	} else {
		fmt.Printf("  err:    %s\n", err)
	}
	a, err = docker.Parse(resourceURl)
	if err == nil {
		fmt.Printf("  gen:   %s\n", a.String())
		if t, ok := a.(docker.Tagged); ok {
			fmt.Printf("  tag:    %s\n", t.Tag())
		}
		if t, ok := a.(docker.Digested); ok {
			fmt.Printf("  digest:  %s\n", t.Digest())
		}
	} else {
		fmt.Printf("  err:    %s\n", err)
	}

}

func repotest() {
	setupCredentials()

	ctx := oci.DefaultContext()

	//spec := ocireg.NewRepositorySpec("ghcr.io")
	//name := "mandelsoft/cnudie/component-descriptors/github.com/mandelsoft/pause"
	version := "0.1-dev"
	spec := docker2.NewRepositorySpec()
	name := "ghcr.io/mandelsoft/pause"

	repo, err := ctx.RepositoryForSpec(spec)
	handleError(err, "get repo")

	ns, err := repo.LookupNamespace(name)
	handleError(err, "lookup namespace")

	art, err := ns.GetArtefact(version)
	handleError(err, "lookup artefact")

	defer art.Close()
	fmt.Println(ociutils.PrintArtefact(art))
	_ = art

	blob := accessio.BlobAccessForString(MIME_OCTET, "testdata")
	err = ns.AddBlob(blob)
	handleError(err, "add blob")

	bd, err := ns.GetBlobData(blob.Digest())
	data, err := bd.Get()
	handleError(err, "read blob")
	fmt.Println(string(data))

	b, _ := art.Blob()
	err = ns.AddTags(b.Digest(), "0.1.beta")
	handleError(err, "add tag")
}

var pattern = regexp.MustCompile("^[0-9a-f]{12}$")

func main() {

	fmt.Printf("%t\n", pattern.MatchString("c4c442d0040d"))
	fmt.Printf("%t\n", !pattern.MatchString("c4c442d0040x"))
	fmt.Printf("%t\n", !pattern.MatchString("c4c442d0040"))
	fmt.Printf("%t\n", !pattern.MatchString("c4c442d0040dd"))

	ref, err := daemon.ParseReference("c4c442d0040d")
	ref, err = daemon.ParseReference("test/laber/blob:latest")
	fmt.Printf("%s\n", ref.StringWithinTransport())
	_ = ref
	_ = err
	Print("ubuntu")
	Print("ubuntu:v1")
	Print("test/ubuntu")
	Print("test/ubuntu:v1")
	Print("ghcr.io/test/ubuntu")
	Print("ghcr.io:8080/test/ubuntu")
	Print("ghcr.io/test/ubuntu:v1")
	Print("ghcr.io/test/ubuntu@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	Print("ghcr.io/test/ubuntu:v1@sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")

	//dockertest()
	dockerwritetest()
	//repotest()

}
