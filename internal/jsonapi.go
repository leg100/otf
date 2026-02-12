package internal

import "github.com/DataDog/jsonapi"

// JSON:API spec forbids '@' in member names. But in OTF it is
// legitimate, such as when terraform state contains an output with map
// value that has '@' in a key:
//
//	output "test" {
//	  description = "test"
//	  value = {
//	    "test.asdf@test" = "asdf"
//	  }
//	}
//
//	By default the upstream JSON:API lib deems it invalid and throws an error.
//	It does so when the agent - which talks to otfd using JSON:API - retrieves
//	the current state from otfd.
//
//	Therefore we disable this validation.
var (
	DefaultJSONAPIMarshalOptions = []jsonapi.MarshalOption{
		jsonapi.MarshalSetNameValidation(jsonapi.DisableValidation),
	}
	DefaultJSONAPIUnmarshalOptions = []jsonapi.UnmarshalOption{
		jsonapi.UnmarshalSetNameValidation(jsonapi.DisableValidation),
	}
)
