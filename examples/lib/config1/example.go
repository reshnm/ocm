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
	"fmt"

	"github.com/open-component-model/ocm/examples/lib/helper"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ccfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/directcreds"
	ociid "github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
)

func UsingConfigs() error {
	cfg, err := helper.ReadConfig(CFG)
	if err != nil {
		return err
	}

	cid := credentials.ConsumerIdentity{
		ociid.ID_TYPE:       ociid.CONSUMER_TYPE,
		ociid.ID_HOSTNAME:   "ghcr.io",
		ociid.ID_PATHPREFIX: "mandelsoft",
	}

	octx := ocm.DefaultContext()
	cctx := octx.ConfigContext()

	// create a credential configuration object
	// and configure it to provide some direct consumer credentials.
	creds := ccfg.New()
	creds.AddConsumer(
		cid,
		directcreds.NewRepositorySpec(cfg.GetCredentials().Properties()),
	)

	err = cctx.ApplyConfig(creds, "explicit")
	if err != nil {
		return errors.Wrapf(err, "cannot apply config")
	}

	credctx := octx.CredentialsContext()

	found, err := credctx.GetCredentialsForConsumer(cid, ociid.IdentityMatcher)
	if err != nil {
		return errors.Wrapf(err, "cannot extract credentials")
	}
	got, err := found.Credentials(credctx)
	if err != nil {
		return errors.Wrapf(err, "cannot evaluate credentials")
	}

	fmt.Printf("found: %s\n", got)
	return nil
}