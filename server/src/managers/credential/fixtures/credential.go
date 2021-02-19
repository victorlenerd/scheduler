package fixtures

import (
	"github.com/bxcodec/faker/v3"
	"scheduler0/server/src/transformers"
	"scheduler0/server/src/utils"
)

type CredentialFixture struct {
	UUID                    string `faker:"uuid_hyphenated"`
	ApiKey                  string `faker:"ipv6"`
	HTTPReferrerRestriction string `faker:"http_referrer_restriction"`
}

func (credentialFixture *CredentialFixture) CreateNCredentialTransformer(n int) []transformers.Credential {
	credentialTransformers := []transformers.Credential{}

	for i := 0; i < n; i++ {
		err := faker.FakeData(credentialFixture)
		utils.CheckErr(err)

		credentialTransformer := transformers.Credential{
			ApiKey:                  credentialFixture.ApiKey,
			HTTPReferrerRestriction: credentialFixture.HTTPReferrerRestriction,
		}

		credentialTransformers = append(credentialTransformers, credentialTransformer)
	}

	return credentialTransformers
}
