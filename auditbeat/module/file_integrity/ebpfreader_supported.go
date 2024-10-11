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

//go:build linux && (amd64 || arm64)

package file_integrity

import "github.com/elastic/elastic-agent-libs/logp"

func newEBPFReader(c Config, l *logp.Logger) (EventProducer, error) {
	paths := make(map[string]struct{})
	for _, p := range c.Paths {
		paths[p] = struct{}{}
	}

	return &ebpfReader{
		config:  c,
		log:     l,
		parsers: FileParsers(c),
		paths:   paths,
		eventC:  make(chan Event),
	}, nil
}
