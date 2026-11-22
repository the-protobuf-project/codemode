// Copyright 2026 The Protobuf Project authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "strings"

// connectorFunc is the generated connector factory: "libraryServiceConnector".
func (s *Service) connectorFunc() string {
	return lowerFirstWord(s.Name) + "Connector"
}

// revertView is a compensating call in connector.ts: the revert RPC, with
// its request built from the original request and response.
type revertView struct {
	Request, Response, Client, Arg string
}

func newRevertView(n moduleNamer, r *Revert) *revertView {
	v := &revertView{Request: "_request", Response: "_response", Client: r.Method.Client()}
	var fields []string
	for _, rf := range r.Fields {
		root := "request"
		if rf.Source == "result" {
			root = "response"
			v.Response = "response"
		} else {
			v.Request = "request"
		}
		fields = append(fields, propertyKey(rf.Field.JSONName)+": "+root+"."+strings.Join(rf.Path, "?."))
	}
	reqType := refType(r.Method.Input, n)
	v.Arg = "{} as " + reqType
	if len(fields) > 0 {
		v.Arg = "{ " + strings.Join(fields, ", ") + " } as " + reqType
	}
	return v
}
