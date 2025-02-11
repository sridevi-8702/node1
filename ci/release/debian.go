/*
 * Copyright (C) 2019 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package release

import (
	"strings"

	"github.com/mysteriumnetwork/go-ci/env"
	"github.com/mysteriumnetwork/go-ci/job"
	"github.com/mysteriumnetwork/go-ci/shell"
	"github.com/mysteriumnetwork/node/ci/deb"
	"github.com/mysteriumnetwork/node/logconfig"
)

type releaseDebianOpts struct {
	repository  string
	version     string
	buildNumber string
}

func releaseDebianPPA(opts *releaseDebianOpts) error {
	if err := deb.TermsTemplateFile("debian/templates"); err != nil {
		return err
	}

	for _, codename := range []string{"bionic", "focal", "jammy", "lunar"} {
		err := shell.NewCmdf("bin/release_ppa %s %s %s %s", opts.repository, opts.version, opts.buildNumber, codename).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func ppaVersion(buildVersion string) string {
	// Using e.g. 0.10.0-rc makes it >= than actual release 0.10.0
	// W: myst source: rc-version-greater-than-expected-version 0.10.0+rc2+build71088039+bionic > 0.10.0
	// (consider using 0.10.0~rc2+build71088039+bionic)
	ver := strings.Replace(buildVersion, "-rc", "~rc", -1)
	// Using ~rc makes it >= ~rc2, so always use a number suffix, starting from rc1
	if strings.HasSuffix(ver, "~rc") {
		ver = strings.Replace(ver, "~rc", "~rc1", -1)
	}
	// PPA treats minus as previous version
	ver = strings.Replace(ver, "-", "+", -1)
	return ver
}

// ReleaseDebianPPASnapshot releases to node-dev PPA
func ReleaseDebianPPASnapshot() error {
	logconfig.Bootstrap()

	if err := env.EnsureEnvVars(
		env.SnapshotBuild,
		env.BuildVersion,
		env.BuildNumber,
	); err != nil {
		return err
	}
	job.Precondition(func() bool {
		return env.Bool(env.SnapshotBuild)
	})

	return releaseDebianPPA(&releaseDebianOpts{
		repository:  "node-dev",
		version:     ppaVersion(env.Str(env.BuildVersion)),
		buildNumber: env.Str(env.BuildNumber),
	})
}

// ReleaseDebianPPAPreRelease releases to node-pre PPA (which is then manually promoted to node PPA)
func ReleaseDebianPPAPreRelease() error {
	logconfig.Bootstrap()

	if err := env.EnsureEnvVars(
		env.TagBuild,
		env.RCBuild,
		env.BuildVersion,
		env.BuildNumber,
	); err != nil {
		return err
	}
	job.Precondition(func() bool {
		return env.Bool(env.TagBuild) && !env.Bool(env.RCBuild)
	})

	return releaseDebianPPA(&releaseDebianOpts{
		repository:  "node-pre",
		version:     ppaVersion(env.Str(env.BuildVersion)),
		buildNumber: env.Str(env.BuildNumber),
	})
}
