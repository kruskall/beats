// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build windows

package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	// errorNames is mapping of errno values to names.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681383(v=vs.85).aspx
	errorNames = map[uint32]string{
		1077: "ERROR_SERVICE_NEVER_STARTED",
	}
)

type Reader struct {
	handle            windows.Handle
	state             uint32
	guid              string            // Host's MachineGuid value (a unique ID for the host).
	ids               map[string]string // Cache of service IDs.
	protectedServices map[string]struct{}
}

func NewReader() (*Reader, error) {
	handle, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ENUMERATE_SERVICE|windows.SC_MANAGER_CONNECT)
	if err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}

	guid, err := getMachineGUID()
	if err != nil {
		return nil, err
	}

	r := &Reader{
		handle:            handle,
		state:             windows.SERVICE_STATE_ALL,
		guid:              guid,
		ids:               map[string]string{},
		protectedServices: map[string]struct{}{},
	}

	return r, nil
}

func (reader *Reader) Read() ([]mapstr.M, error) {
	services, err := GetServiceStates(reader.handle, reader.state, reader.protectedServices)
	if err != nil {
		return nil, err
	}

	result := make([]mapstr.M, 0, len(services))

	for _, service := range services {
		ev := mapstr.M{
			"id":           reader.getServiceID(service.ServiceName),
			"display_name": service.DisplayName,
			"name":         service.ServiceName,
			"state":        service.CurrentState,
			"start_type":   service.StartType.String(),
			"start_name":   service.ServiceStartName,
			"path_name":    service.BinaryPathName,
		}

		if service.CurrentState == "Stopped" {
			ev.Put("exit_code", getErrorCode(service.ExitCode))
		}

		if service.PID > 0 {
			ev.Put("pid", service.PID)
		}

		if service.Uptime > 0 {
			if _, err = ev.Put("uptime.ms", service.Uptime); err != nil {
				return nil, err
			}
		}

		result = append(result, ev)
	}

	return result, nil
}

func (reader *Reader) Close() error {
	return windows.CloseServiceHandle(reader.handle)
}

// getMachineGUID returns the machine's GUID value which is unique to a Windows
// installation.
func getMachineGUID() (string, error) {
	const key = registry.LOCAL_MACHINE
	const path = `SOFTWARE\Microsoft\Cryptography`
	const name = "MachineGuid"

	k, err := registry.OpenKey(key, path, registry.READ|registry.WOW64_64KEY)
	if err != nil {
		return "", fmt.Errorf(`failed to open HKLM\%v: %w`, path, err)
	}

	guid, _, err := k.GetStringValue(name)
	if err != nil {
		return "", fmt.Errorf(`failed to get value of HKLM\%v\%v: %w`, path, name, err)
	}

	return guid, nil
}

// getServiceID returns a unique ID for the service that is derived from the
// machine's GUID and the service's name.
func (reader *Reader) getServiceID(name string) string {
	// hash returns a base64 encoded sha256 hash that is truncated to 10 chars.
	hash := func(v string) string {
		sum := sha256.Sum256([]byte(v))
		base64Hash := base64.RawURLEncoding.EncodeToString(sum[:])
		return base64Hash[:10]
	}

	id, found := reader.ids[name]
	if !found {
		id = hash(reader.guid + name)
		reader.ids[name] = id
	}

	return id
}

func getErrorCode(errno uint32) string {
	name, found := errorNames[errno]
	if found {
		return name
	}
	return strconv.Itoa(int(errno))
}
