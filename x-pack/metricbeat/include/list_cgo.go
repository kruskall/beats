// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

//go:build cgo

package include

import (
	_ "github.com/elastic/beats/v7/x-pack/metricbeat/module/oracle"
	_ "github.com/elastic/beats/v7/x-pack/metricbeat/module/oracle/performance"
	_ "github.com/elastic/beats/v7/x-pack/metricbeat/module/oracle/sysmetric"
	_ "github.com/elastic/beats/v7/x-pack/metricbeat/module/oracle/tablespace"
)
