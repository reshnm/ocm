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

package output

import (
	"context"

	. "github.com/gardener/ocm/cmds/ocm/pkg/data"
)

type ElementOutput struct {
	source ProcessingSource
	Elems  Iterable
}

func NewElementOutput(chain ProcessChain) *ElementOutput {
	return (&ElementOutput{}).new(chain)
}

func (this *ElementOutput) new(chain ProcessChain) *ElementOutput {
	this.source = NewIncrementalProcessingSource()
	if chain == nil {
		this.Elems = this.source
	} else {
		this.Elems = Process(this.source).Asynchronously().Apply(chain)
	}
	return this
}

func (this *ElementOutput) Add(ctx *context.Context, e interface{}) error {
	this.source.Add(e)
	return nil
}

func (this *ElementOutput) Close(ctx *context.Context) error {
	this.source.Close()
	return nil
}

func (this *ElementOutput) Out(ctx *context.Context) {
}