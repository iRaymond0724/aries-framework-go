/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vcwallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	vdrapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/internal/jsonldtest"
	mockprovider "github.com/hyperledger/aries-framework-go/pkg/mock/provider"
	mockstorage "github.com/hyperledger/aries-framework-go/pkg/mock/storage"
	mockvdr "github.com/hyperledger/aries-framework-go/pkg/mock/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/wallet"
)

// nolint: lll
const (
	sampleUserID           = "sample-user01"
	samplePassPhrase       = "fakepassphrase"
	sampleKeyStoreURL      = "sample/keyserver/test"
	sampleEDVServerURL     = "sample-edv-url"
	sampleEDVVaultID       = "sample-edv-vault-id"
	sampleEDVEncryptionKID = "sample-edv-encryption-kid"
	sampleEDVMacKID        = "sample-edv-mac-kid"
	sampleCommandError     = "sample-command-error-01"
	sampleFakeTkn          = "sample-fake-token-01"
	sampleDIDKey           = "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
	sampleUDCVC            = `{
      "@context": [
        "https://www.w3.org/2018/credentials/v1",
        "https://www.w3.org/2018/credentials/examples/v1"
      ],
     "credentialSchema": [],
      "credentialSubject": {
        "degree": {
          "type": "BachelorDegree",
          "university": "MIT"
        },
        "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
        "name": "Jayden Doe",
        "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
      },
      "expirationDate": "2020-01-01T19:23:24Z",
      "id": "http://example.edu/credentials/1877",
      "issuanceDate": "2010-01-01T19:23:24Z",
      "issuer": {
        "id": "did:example:76e12ec712ebc6f1c221ebfeb1f",
        "name": "Example University"
      },
      "referenceNumber": 83294847,
      "type": [
        "VerifiableCredential",
        "UniversityDegreeCredential"
      ]
    }`
	sampleMetadata = `{
        "@context": ["https://w3id.org/wallet/v1"],
        "id": "urn:uuid:2905324a-9524-11ea-bb37-0242ac130002",
        "type": "Metadata",
        "name": "Ropsten Testnet HD Accounts",
        "image": "https://via.placeholder.com/150",
        "description": "My Ethereum TestNet Accounts",
        "tags": ["professional", "organization"],
        "correlation": ["urn:uuid:4058a72a-9523-11ea-bb37-0242ac130002"],
        "hdPath": "m’/44’/60’/0’",
        "target": ["urn:uuid:c410e44a-9525-11ea-bb37-0242ac130002"]
    }`
	sampleBBSVC = `{
            "@context": ["https://www.w3.org/2018/credentials/v1", "https://www.w3.org/2018/credentials/examples/v1", "https://w3id.org/security/bbs/v1"],
            "credentialSubject": {
                "degree": {"type": "BachelorDegree", "university": "MIT"},
                "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
                "name": "Jayden Doe",
                "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
            },
            "expirationDate": "2020-01-01T19:23:24Z",
            "id": "http://example.edu/credentials/1872",
            "issuanceDate": "2010-01-01T19:23:24Z",
            "issuer": {"id": "did:example:76e12ec712ebc6f1c221ebfeb1f", "name": "Example University"},
            "proof": {
                "created": "2021-03-29T13:27:36.483097-04:00",
                "proofPurpose": "assertionMethod",
                "proofValue": "rw7FeV6K1wimnYogF9qd-N0zmq5QlaIoszg64HciTca-mK_WU4E1jIusKTT6EnN2GZz04NVPBIw4yhc0kTwIZ07etMvfWUlHt_KMoy2CfTw8FBhrf66q4h7Qcqxh_Kxp6yCHyB4A-MmURlKKb8o-4w",
                "type": "BbsBlsSignature2020",
                "verificationMethod": "did:key:zUC72c7u4BYVmfYinDceXkNAwzPEyuEE23kUmJDjLy8495KH3pjLwFhae1Fww9qxxRdLnS2VNNwni6W3KbYZKsicDtiNNEp76fYWR6HCD8jAz6ihwmLRjcHH6kB294Xfg1SL1qQ#zUC72c7u4BYVmfYinDceXkNAwzPEyuEE23kUmJDjLy8495KH3pjLwFhae1Fww9qxxRdLnS2VNNwni6W3KbYZKsicDtiNNEp76fYWR6HCD8jAz6ihwmLRjcHH6kB294Xfg1SL1qQ"
            },
            "referenceNumber": 83294847,
            "type": ["VerifiableCredential", "UniversityDegreeCredential"]
        }`
	sampleQueryByExample = `{
                        "reason": "Please present your identity document.",
                        "example": {
                            "@context": [
								"https://www.w3.org/2018/credentials/v1",
								"https://www.w3.org/2018/credentials/examples/v1"
                            ],
                            "type": ["UniversityDegreeCredential"],
							"trustedIssuer": [
              					{
                					"issuer": "urn:some:required:issuer"
              					},
								{
                					"required": true,
                					"issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f"
              					}
							],
							"credentialSubject": {
								"id": "did:example:ebfeb1f712ebc6f1c276e12ec21"	
							}
                        }
                	}`
	sampleQueryByFrame = `{
                    "reason": "Please provide your Passport details.",
                    "frame": {
                        "@context": [
                            "https://www.w3.org/2018/credentials/v1",
                            "https://w3id.org/citizenship/v1",
                            "https://w3id.org/security/bbs/v1"
                        ],
                        "type": ["VerifiableCredential", "PermanentResidentCard"],
                        "@explicit": true,
                        "identifier": {},
                        "issuer": {},
                        "issuanceDate": {},
                        "credentialSubject": {
                            "@explicit": true,
                            "name": {},
                            "spouse": {}
                        }
                    },
                    "trustedIssuer": [
                        {
                            "issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f",
                            "required": true
                        }
                    ],
                    "required": true
                }`
	sampleFrame = `
		{
			"@context": [
				"https://www.w3.org/2018/credentials/v1",
				"https://www.w3.org/2018/credentials/examples/v1",
				"https://w3id.org/security/bbs/v1"
			],
  			"type": ["VerifiableCredential", "UniversityDegreeCredential"],
  			"@explicit": true,
  			"identifier": {},
  			"issuer": {},
  			"issuanceDate": {},
  			"credentialSubject": {
    			"@explicit": true,
    			"degree": {},
    			"name": {}
  			}
		}
	`
	sampleKeyContentBase58 = `{
  			"@context": ["https://w3id.org/wallet/v1"],
  		  	"id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
  		  	"controller": "did:example:123456789abcdefghi",
			"type": "Ed25519VerificationKey2018",
			"privateKeyBase58":"2MP5gWCnf67jvW3E4Lz8PpVrDWAXMYY1sDxjnkEnKhkkbKD7yP2mkVeyVpu5nAtr3TeDgMNjBPirk2XcQacs3dvZ"
  		}`
	sampleDIDResolutionResponse = `{
        "@context": [
            "https://w3id.org/wallet/v1",
            "https://w3id.org/did-resolution/v1"
        ],
        "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
        "type": ["DIDResolutionResponse"],
        "name": "Farming Sensor DID Document",
        "image": "https://via.placeholder.com/150",
        "description": "An IoT device in the middle of a corn field.",
        "tags": ["professional"],
        "correlation": ["4058a72a-9523-11ea-bb37-0242ac130002"],
        "created": "2017-06-18T21:19:10Z",
        "expires": "2026-06-18T21:19:10Z",
        "didDocument": {
            "@context": [
                "https://w3id.org/did/v0.11"
            ],
            "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
            "publicKey": [
                {
                    "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "type": "Ed25519VerificationKey2018",
                    "controller": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "publicKeyBase58": "8jkuMBqmu1TRA6is7TT5tKBksTZamrLhaXrg9NAczqeh"
                }
            ],
            "authentication": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "assertionMethod": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "capabilityDelegation": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "capabilityInvocation": [
                "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5"
            ],
            "keyAgreement": [
                {
                    "id": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6LSmjNfS5FC9W59JtPZq7fHgrjThxsidjEhZeMxCarbR998",
                    "type": "X25519KeyAgreementKey2019",
                    "controller": "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
                    "publicKeyBase58": "B4CVumSL43MQDW1oJU9LNGWyrpLbw84YgfeGi8D4hmNN"
                }
            ]
        }
    }`
)

func TestNew(t *testing.T) {
	t.Run("successfully create new command instance", func(t *testing.T) {
		cmd := New(newMockProvider(t))
		require.NotNil(t, cmd)

		require.Len(t, cmd.GetHandlers(), 13)
	})
}

func TestCommand_CreateProfile(t *testing.T) {
	t.Run("successfully create a new wallet profile (localkms)", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		request := &CreateOrUpdateProfileRequest{
			UserID:             sampleUserID,
			LocalKMSPassphrase: samplePassPhrase,
		}

		var b bytes.Buffer
		cmdErr := cmd.CreateProfile(&b, getReader(t, &request))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err := wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)
	})

	t.Run("successfully create a new wallet profile (webkms/remotekms)", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		request := &CreateOrUpdateProfileRequest{
			UserID:      sampleUserID,
			KeyStoreURL: sampleKeyStoreURL,
		}

		var b bytes.Buffer
		cmdErr := cmd.CreateProfile(&b, getReader(t, &request))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err := wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)
	})

	t.Run("successfully create a new wallet profile with EDV configuration", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		// create with remote kms.
		request := &CreateOrUpdateProfileRequest{
			UserID:      uuid.New().String(),
			KeyStoreURL: sampleKeyStoreURL,
			EDVConfiguration: &EDVConfiguration{
				ServerURL:       sampleEDVServerURL,
				VaultID:         sampleEDVVaultID,
				MACKeyID:        sampleEDVMacKID,
				EncryptionKeyID: sampleEDVEncryptionKID,
			},
		}

		var b1 bytes.Buffer
		cmdErr := cmd.CreateProfile(&b1, getReader(t, &request))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err := wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)

		// create with local kms.
		request = &CreateOrUpdateProfileRequest{
			UserID:             uuid.New().String(),
			LocalKMSPassphrase: samplePassPhrase,
			EDVConfiguration: &EDVConfiguration{
				ServerURL: sampleEDVServerURL,
				VaultID:   sampleEDVVaultID,
			},
		}

		var b2 bytes.Buffer
		cmdErr = cmd.CreateProfile(&b2, getReader(t, &request))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err = wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)
	})

	t.Run("failed to create duplicate profile", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		request := &CreateOrUpdateProfileRequest{
			UserID:             sampleUserID,
			LocalKMSPassphrase: samplePassPhrase,
		}

		var b1 bytes.Buffer
		cmdErr := cmd.CreateProfile(&b1, getReader(t, &request))
		require.NoError(t, cmdErr)

		request = &CreateOrUpdateProfileRequest{
			UserID:      sampleUserID,
			KeyStoreURL: sampleKeyStoreURL,
		}

		var b2 bytes.Buffer
		cmdErr = cmd.CreateProfile(&b2, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Equal(t, cmdErr.Code(), CreateProfileErrorCode)
	})

	t.Run("failed to create profile due to invalid settings", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		request := &CreateOrUpdateProfileRequest{
			UserID: sampleUserID,
		}

		var b1 bytes.Buffer
		cmdErr := cmd.CreateProfile(&b1, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Code(), CreateProfileErrorCode)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
	})

	t.Run("failed to create profile due to invalid request", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		var b1 bytes.Buffer
		cmdErr := cmd.CreateProfile(&b1, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Code(), InvalidRequestErrorCode)
		require.Equal(t, cmdErr.Type(), command.ValidationError)
	})

	t.Run("failed to create profile due to EDV key set creation failure", func(t *testing.T) {
		mockctx := newMockProvider(t)

		cmd := New(mockctx)
		require.NotNil(t, cmd)

		mockStProv, ok := mockctx.StorageProviderValue.(*mockstorage.MockStoreProvider)
		require.True(t, ok)
		require.NotEmpty(t, mockStProv)

		mockStProv.Store.ErrGet = errors.New(sampleCommandError)

		request := &CreateOrUpdateProfileRequest{
			UserID:             uuid.New().String(),
			LocalKMSPassphrase: samplePassPhrase,
			EDVConfiguration: &EDVConfiguration{
				ServerURL: sampleEDVServerURL,
				VaultID:   sampleEDVVaultID,
			},
		}

		var b1 bytes.Buffer
		cmdErr := cmd.CreateProfile(&b1, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Code(), CreateProfileErrorCode)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
		require.Contains(t, cmdErr.Error(), sampleCommandError)
	})
}

func TestCommand_UpdateProfile(t *testing.T) {
	mockctx := newMockProvider(t)

	cmd := New(mockctx)
	require.NotNil(t, cmd)

	createRqst := &CreateOrUpdateProfileRequest{
		UserID:             sampleUserID,
		LocalKMSPassphrase: samplePassPhrase,
	}

	var c bytes.Buffer
	cmdErr := cmd.CreateProfile(&c, getReader(t, &createRqst))
	require.NoError(t, cmdErr)

	t.Run("successfully update a wallet profile", func(t *testing.T) {
		request := &CreateOrUpdateProfileRequest{
			UserID:      sampleUserID,
			KeyStoreURL: sampleKeyStoreURL,
		}

		var b bytes.Buffer
		cmdErr := cmd.UpdateProfile(&b, getReader(t, &createRqst))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err := wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)
	})

	t.Run("successfully update a wallet profile with EDV configuration", func(t *testing.T) {
		// create with remote kms.
		request := &CreateOrUpdateProfileRequest{
			UserID:      sampleUserID,
			KeyStoreURL: sampleKeyStoreURL,
			EDVConfiguration: &EDVConfiguration{
				ServerURL:       sampleEDVServerURL,
				VaultID:         sampleEDVVaultID,
				MACKeyID:        sampleEDVMacKID,
				EncryptionKeyID: sampleEDVEncryptionKID,
			},
		}

		var b1 bytes.Buffer
		cmdErr := cmd.UpdateProfile(&b1, getReader(t, &request))
		require.NoError(t, cmdErr)

		// if wallet instance can be creates it means profile exists
		walletInstance, err := wallet.New(request.UserID, mockctx)
		require.NoError(t, err)
		require.NotEmpty(t, walletInstance)
	})

	t.Run("failed to update profile due to invalid settings", func(t *testing.T) {
		request := &CreateOrUpdateProfileRequest{
			UserID: sampleUserID,
		}

		var b1 bytes.Buffer
		cmdErr := cmd.UpdateProfile(&b1, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Code(), UpdateProfileErrorCode)
		require.Equal(t, cmdErr.Type(), command.ExecuteError)
	})

	t.Run("failed to update profile due to invalid request", func(t *testing.T) {
		var b1 bytes.Buffer
		cmdErr := cmd.UpdateProfile(&b1, bytes.NewBufferString("--"))
		require.Error(t, cmdErr)
		require.Equal(t, cmdErr.Code(), InvalidRequestErrorCode)
		require.Equal(t, cmdErr.Type(), command.ValidationError)
	})
}

func TestCommand_OpenAndClose(t *testing.T) {
	const (
		sampleUser1 = "sample-user-01"
		sampleUser2 = "sample-user-02"
		sampleUser3 = "sample-user-03"
	)

	mockctx := newMockProvider(t)

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:      sampleUser2,
		KeyStoreURL: sampleKeyStoreURL,
	})

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser3,
		LocalKMSPassphrase: samplePassPhrase,
		EDVConfiguration: &EDVConfiguration{
			ServerURL: sampleEDVServerURL,
			VaultID:   sampleEDVVaultID,
		},
	})

	t.Run("successfully unlock & lock wallet (local kms)", func(t *testing.T) {
		cmd := New(mockctx)

		request := &UnlockWalletRequest{
			UserID:             sampleUser1,
			LocalKMSPassphrase: samplePassPhrase,
		}

		// unlock wallet
		var b bytes.Buffer
		cmdErr := cmd.Open(&b, getReader(t, &request))
		require.NoError(t, cmdErr)
		require.NotEmpty(t, getUnlockToken(t, b))
		b.Reset()

		// try again, should get error, wallet already unlocked
		cmdErr = cmd.Open(&b, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), wallet.ErrAlreadyUnlocked.Error())
		require.Empty(t, b.Len())
		b.Reset()

		// lock wallet
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser1}))
		require.NoError(t, cmdErr)
		var lockResponse LockWalletResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.True(t, lockResponse.Closed)
		b.Reset()

		// lock wallet again
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser1}))
		require.NoError(t, cmdErr)
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.False(t, lockResponse.Closed)
		b.Reset()
	})

	t.Run("successfully unlock & lock wallet (remote kms)", func(t *testing.T) {
		cmd := New(mockctx)

		request := &UnlockWalletRequest{
			UserID:     sampleUser2,
			WebKMSAuth: sampleFakeTkn,
		}

		// unlock wallet
		var b bytes.Buffer
		cmdErr := cmd.Open(&b, getReader(t, &request))
		require.NoError(t, cmdErr)
		require.NotEmpty(t, getUnlockToken(t, b))
		b.Reset()

		// try again, should get error, wallet already unlocked
		cmdErr = cmd.Open(&b, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), wallet.ErrAlreadyUnlocked.Error())
		require.Empty(t, b.Len())
		b.Reset()

		// lock wallet
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser2}))
		require.NoError(t, cmdErr)
		var lockResponse LockWalletResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.True(t, lockResponse.Closed)
		b.Reset()

		// lock wallet again
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser2}))
		require.NoError(t, cmdErr)
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.False(t, lockResponse.Closed)
		b.Reset()
	})

	t.Run("successfully unlock & lock wallet (local kms, edv user)", func(t *testing.T) {
		cmd := New(mockctx)

		request := &UnlockWalletRequest{
			UserID:             sampleUser3,
			LocalKMSPassphrase: samplePassPhrase,
			EDVUnlock: &EDVUnlockRequest{
				AuthToken: sampleFakeTkn,
			},
		}

		// unlock wallet
		var b bytes.Buffer
		cmdErr := cmd.Open(&b, getReader(t, &request))
		require.NoError(t, cmdErr)
		require.NotEmpty(t, getUnlockToken(t, b))
		b.Reset()

		// try again, should get error, wallet already unlocked
		cmdErr = cmd.Open(&b, getReader(t, &request))
		require.Error(t, cmdErr)
		require.Contains(t, cmdErr.Error(), wallet.ErrAlreadyUnlocked.Error())
		require.Empty(t, b.Len())
		b.Reset()

		// lock wallet
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser3}))
		require.NoError(t, cmdErr)
		var lockResponse LockWalletResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.True(t, lockResponse.Closed)
		b.Reset()

		// lock wallet again
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: sampleUser3}))
		require.NoError(t, cmdErr)
		require.NoError(t, json.NewDecoder(&b).Decode(&lockResponse))
		require.False(t, lockResponse.Closed)
		b.Reset()
	})

	t.Run("lock & unlock failures", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Open(&b, getReader(t, &UnlockWalletRequest{}))
		require.Error(t, cmdErr)
		validateError(t, cmdErr, command.ExecuteError, OpenWalletErrorCode, "profile does not exist")
		require.Empty(t, b.Len())
		b.Reset()

		cmdErr = cmd.Open(&b, getReader(t, ""))
		require.Error(t, cmdErr)
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, "cannot unmarshal string into Go")
		require.Empty(t, b.Len())
		b.Reset()

		cmdErr = cmd.Close(&b, getReader(t, &UnlockWalletRequest{}))
		require.Error(t, cmdErr)
		validateError(t, cmdErr, command.ExecuteError, CloseWalletErrorCode, "profile does not exist")
		require.Empty(t, b.Len())
		b.Reset()

		cmdErr = cmd.Close(&b, getReader(t, ""))
		require.Error(t, cmdErr)
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, "cannot unmarshal string into Go")
		require.Empty(t, b.Len())
		b.Reset()
	})
}

func TestCommand_AddRemoveGetGetAll(t *testing.T) {
	const (
		sampleUser1 = "sample-user-01"
		sampleUser2 = "sample-user-02"
		sampleUser3 = "sample-user-03"
	)

	mockctx := newMockProvider(t)

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	token1, lock1 := unlockWallet(t, mockctx, &UnlockWalletRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	defer lock1()

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:      sampleUser2,
		KeyStoreURL: sampleKeyStoreURL,
	})

	token2, lock2 := unlockWallet(t, mockctx, &UnlockWalletRequest{
		UserID:     sampleUser2,
		WebKMSAuth: sampleFakeTkn,
	})

	defer lock2()

	t.Run("add a credential to wallet", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
			Content:     []byte(sampleUDCVC),
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)
	})

	t.Run("add a metadata to wallet", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
			Content:     []byte(sampleMetadata),
			ContentType: "metadata",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)
	})

	t.Run("get a credential from wallet", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Get(&b, getReader(t, &GetContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)

		var response GetContentResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.NotEmpty(t, response)
		require.NotEmpty(t, response.Content)
	})

	t.Run("get all credentials from wallet", func(t *testing.T) {
		cmd := New(mockctx)

		// save multiple credentials, one already saved
		const count = 6
		for i := 1; i < count; i++ {
			var b bytes.Buffer
			cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
				Content: []byte(strings.ReplaceAll(sampleUDCVC, `"http://example.edu/credentials/1877"`,
					fmt.Sprintf(`"http://example.edu/credentials/1872%d"`, i))),
				ContentType: "credential",
				WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
			}))
			require.NoError(t, cmdErr)

			b.Reset()
		}

		var b bytes.Buffer

		cmdErr := cmd.GetAll(&b, getReader(t, &GetAllContentRequest{
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)

		var response GetAllContentResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.NotEmpty(t, response)
		require.Len(t, response.Contents, count)
	})

	t.Run("remove a credential from wallet", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Remove(&b, getReader(t, &RemoveContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)
	})

	t.Run("get a credential from different wallet", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Get(&b, getReader(t, &GetContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser2, Auth: token2},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetFromWalletErrorCode, "data not found")
	})

	t.Run("try content operations from invalid auth", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		const expectedErr = "invalid auth token"

		cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
			Content:     []byte(sampleUDCVC),
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, AddToWalletErrorCode, expectedErr)
		b.Reset()

		cmdErr = cmd.Get(&b, getReader(t, &GetContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetFromWalletErrorCode, expectedErr)

		cmdErr = cmd.GetAll(&b, getReader(t, &GetAllContentRequest{
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetAllFromWalletErrorCode, expectedErr)

		cmdErr = cmd.Remove(&b, getReader(t, &RemoveContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, RemoveFromWalletErrorCode, expectedErr)
	})

	t.Run("try content operations from invalid content type", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
			Content:     []byte(sampleUDCVC),
			ContentType: "mango",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		validateError(t, cmdErr, command.ExecuteError, AddToWalletErrorCode, "invalid content type")
		b.Reset()

		cmdErr = cmd.Get(&b, getReader(t, &GetContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "pineapple",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetFromWalletErrorCode, "data not found")

		cmdErr = cmd.GetAll(&b, getReader(t, &GetAllContentRequest{
			ContentType: "orange",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)

		var response GetAllContentResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.Empty(t, response.Contents)
		b.Reset()

		cmdErr = cmd.Remove(&b, getReader(t, &RemoveContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "strawberry",
			WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token1},
		}))
		require.NoError(t, cmdErr)
	})

	t.Run("try content operations from invalid profile", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		const expectedErr = "profile does not exist"

		cmdErr := cmd.Add(&b, getReader(t, &AddContentRequest{
			Content:     []byte(sampleUDCVC),
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser3, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, AddToWalletErrorCode, expectedErr)
		b.Reset()

		cmdErr = cmd.Get(&b, getReader(t, &GetContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser3, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetFromWalletErrorCode, expectedErr)

		cmdErr = cmd.GetAll(&b, getReader(t, &GetAllContentRequest{
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser3, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, GetAllFromWalletErrorCode, expectedErr)

		cmdErr = cmd.Remove(&b, getReader(t, &RemoveContentRequest{
			ContentID:   "http://example.edu/credentials/1877",
			ContentType: "credential",
			WalletAuth:  WalletAuth{UserID: sampleUser3, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, RemoveFromWalletErrorCode, expectedErr)
	})

	t.Run("try content operations from invalid request", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		const expectedErr = "invalid character"

		cmdErr := cmd.Add(&b, bytes.NewBufferString("invalid request"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, expectedErr)
		b.Reset()

		cmdErr = cmd.Get(&b, bytes.NewBufferString("invalid request"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, expectedErr)

		cmdErr = cmd.GetAll(&b, bytes.NewBufferString("invalid request"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, expectedErr)

		cmdErr = cmd.Remove(&b, bytes.NewBufferString("invalid request"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, expectedErr)
	})
}

func TestCommand_Query(t *testing.T) {
	const sampleUser1 = "sample-user-01"

	mockctx := newMockProvider(t)
	mockctx.VDRegistryValue = getMockDIDKeyVDR()

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	token, lock := unlockWallet(t, mockctx, &UnlockWalletRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	defer lock()

	addContent(t, mockctx, &AddContentRequest{
		Content:     []byte(sampleUDCVC),
		ContentType: "credential",
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})

	addContent(t, mockctx, &AddContentRequest{
		Content:     []byte(sampleBBSVC),
		ContentType: "credential",
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})

	t.Run("successfully query credentials", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Query(&b, getReader(t, &ContentQueryRequest{
			Query: []*wallet.QueryParams{
				{
					Type:  "QueryByExample",
					Query: []json.RawMessage{[]byte(sampleQueryByExample)},
				},
				{
					Type:  "QueryByFrame",
					Query: []json.RawMessage{[]byte(sampleQueryByFrame)},
				},
			},
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: token},
		}))
		require.NoError(t, cmdErr)

		var response map[string]interface{}
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.NotEmpty(t, response)
		require.NotEmpty(t, response["results"])
	})

	t.Run("query credentials with invalid auth", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Query(&b, getReader(t, &ContentQueryRequest{
			Query: []*wallet.QueryParams{
				{
					Type:  "QueryByFrame",
					Query: []json.RawMessage{[]byte(sampleQueryByFrame)},
				},
			},
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, QueryWalletErrorCode, "invalid auth token")
	})

	t.Run("query credentials with invalid wallet profile", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Query(&b, getReader(t, &ContentQueryRequest{
			Query: []*wallet.QueryParams{
				{
					Type:  "QueryByFrame",
					Query: []json.RawMessage{[]byte(sampleQueryByFrame)},
				},
			},
			WalletAuth: WalletAuth{UserID: sampleUserID, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ExecuteError, QueryWalletErrorCode, "profile does not exist")
	})

	t.Run("query credentials with invalid query type", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Query(&b, getReader(t, &ContentQueryRequest{
			Query: []*wallet.QueryParams{
				{
					Type:  "QueryByOrange",
					Query: []json.RawMessage{[]byte(sampleQueryByFrame)},
				},
			},
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: token},
		}))
		validateError(t, cmdErr, command.ExecuteError, QueryWalletErrorCode, "unsupported query type")
	})

	t.Run("query credentials with invalid request", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Query(&b, bytes.NewBufferString("--"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, "invalid character")
	})
}

func TestCommand_IssueProveVerify(t *testing.T) {
	const sampleUser1 = "sample-user-01"

	mockctx := newMockProvider(t)
	mockctx.VDRegistryValue = getMockDIDKeyVDR()

	tcrypto, err := tinkcrypto.New()
	require.NoError(t, err)

	mockctx.CryptoValue = tcrypto

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	token, lock := unlockWallet(t, mockctx, &UnlockWalletRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	defer lock()

	addContent(t, mockctx, &AddContentRequest{
		Content:     []byte(sampleKeyContentBase58),
		ContentType: wallet.Key,
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})
	addContent(t, mockctx, &AddContentRequest{
		Content:     []byte(sampleDIDResolutionResponse),
		ContentType: wallet.DIDResolutionResponse,
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})

	var rawCredentialToVerify json.RawMessage

	t.Run("issue a credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Issue(&b, getReader(t, &IssueRequest{
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: token},
			Credential: []byte(sampleUDCVC),
			ProofOptions: &wallet.ProofOptions{
				Controller: sampleDIDKey,
			},
		}))
		require.NoError(t, cmdErr)

		credentialIssued := parseCredential(t, b)
		require.Len(t, credentialIssued.Proofs, 1)
		b.Reset()

		rawCredentialToVerify, err = credentialIssued.MarshalJSON()
		require.NoError(t, err)
	})

	// save it in store for next tests
	addContent(t, mockctx, &AddContentRequest{
		Content:     rawCredentialToVerify,
		ContentType: wallet.Credential,
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})

	t.Run("verify a credential from store", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:         WalletAuth{UserID: sampleUser1, Auth: token},
			StoredCredentialID: "http://example.edu/credentials/1877",
		}))
		require.NoError(t, cmdErr)

		var response VerifyResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.True(t, response.Verified)
		require.Empty(t, response.Error)
	})

	t.Run("verify a raw credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:    WalletAuth{UserID: sampleUser1, Auth: token},
			RawCredential: rawCredentialToVerify,
		}))
		require.NoError(t, cmdErr)

		var response VerifyResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.True(t, response.Verified)
		require.Empty(t, response.Error)
	})

	t.Run("verify a invalid credential", func(t *testing.T) {
		// tamper a credential
		invalidVC := string(rawCredentialToVerify)
		invalidVC = strings.ReplaceAll(invalidVC, "Jayden Doe", "John Smith")

		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:    WalletAuth{UserID: sampleUser1, Auth: token},
			RawCredential: []byte(invalidVC),
		}))
		require.NoError(t, cmdErr)

		var response VerifyResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.False(t, response.Verified)
		require.NotEmpty(t, response.Error)
		require.Contains(t, response.Error, "invalid signature")
	})

	var presentation *verifiable.Presentation

	t.Run("prove credentials", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Prove(&b, getReader(t, &ProveRequest{
			WalletAuth:        WalletAuth{UserID: sampleUser1, Auth: token},
			RawCredentials:    []json.RawMessage{rawCredentialToVerify},
			StoredCredentials: []string{"http://example.edu/credentials/1877"},
			ProofOptions: &wallet.ProofOptions{
				Controller: sampleDIDKey,
			},
		}))
		require.NoError(t, cmdErr)

		presentation = parsePresentation(t, b)
		require.NotEmpty(t, presentation.Proofs)
		require.Len(t, presentation.Credentials(), 2)
		require.Len(t, presentation.Proofs, 1)
		b.Reset()

		// prove using raw presentation
		rawPresentation, err := presentation.MarshalJSON()
		require.NoError(t, err)

		cmdErr = cmd.Prove(&b, getReader(t, &ProveRequest{
			WalletAuth:        WalletAuth{UserID: sampleUser1, Auth: token},
			StoredCredentials: []string{"http://example.edu/credentials/1877"},
			Presentation:      rawPresentation,
			ProofOptions: &wallet.ProofOptions{
				Controller: sampleDIDKey,
			},
		}))
		require.NoError(t, cmdErr)
		presentation2 := parsePresentation(t, b)
		require.NotEmpty(t, presentation2.Proofs)
		require.Len(t, presentation2.Credentials(), 3)
		require.Len(t, presentation2.Proofs, 2)
	})

	t.Run("verify a raw presentation", func(t *testing.T) {
		vpBytes, err := presentation.MarshalJSON()
		require.NoError(t, err)

		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:   WalletAuth{UserID: sampleUser1, Auth: token},
			Presentation: vpBytes,
		}))
		require.NoError(t, cmdErr)

		var response VerifyResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.True(t, response.Verified)
		require.Empty(t, response.Error)
		b.Reset()

		// tamper it and try
		invalidVP := string(vpBytes)
		invalidVP = strings.ReplaceAll(invalidVP, "Jayden Doe", "John Smith")

		cmdErr = cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:   WalletAuth{UserID: sampleUser1, Auth: token},
			Presentation: []byte(invalidVP),
		}))
		require.NoError(t, cmdErr)

		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.False(t, response.Verified)
		require.NotEmpty(t, response.Error)
		require.Contains(t, response.Error, "invalid signature")
		b.Reset()
	})

	t.Run("failed to prove a credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Issue(&b, getReader(t, &IssueRequest{
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: token},
			Credential: []byte(sampleUDCVC),
			ProofOptions: &wallet.ProofOptions{
				Controller: "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv464",
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, IssueFromWalletErrorCode, "failed to prepare proof")
	})

	t.Run("failed to prove a credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Prove(&b, getReader(t, &ProveRequest{
			WalletAuth:        WalletAuth{UserID: sampleUser1, Auth: token},
			RawCredentials:    []json.RawMessage{rawCredentialToVerify},
			StoredCredentials: []string{"http://example.edu/credentials/1877"},
			ProofOptions: &wallet.ProofOptions{
				Controller: "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv464",
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, ProveFromWalletErrorCode, "failed to prepare proof")
	})

	t.Run("issue,prove,verify with invalid profile", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		const errMsg = "profile does not exist"

		cmdErr := cmd.Prove(&b, getReader(t, &ProveRequest{
			WalletAuth:        WalletAuth{UserID: sampleUserID, Auth: token},
			StoredCredentials: []string{"http://example.edu/credentials/1877"},
			ProofOptions: &wallet.ProofOptions{
				Controller: "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv464",
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, ProveFromWalletErrorCode, errMsg)
		b.Reset()

		cmdErr = cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth:         WalletAuth{UserID: sampleUserID, Auth: token},
			StoredCredentialID: "http://example.edu/credentials/1877",
		}))
		validateError(t, cmdErr, command.ExecuteError, VerifyFromWalletErrorCode, errMsg)
		b.Reset()

		cmdErr = cmd.Issue(&b, getReader(t, &IssueRequest{
			WalletAuth: WalletAuth{UserID: sampleUserID, Auth: token},
			Credential: []byte(sampleUDCVC),
			ProofOptions: &wallet.ProofOptions{
				Controller: sampleDIDKey,
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, IssueFromWalletErrorCode, errMsg)
		b.Reset()
	})

	t.Run("issue,prove,verify with invalid auth", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		const errMsg = "invalid auth token"

		cmdErr := cmd.Prove(&b, getReader(t, &ProveRequest{
			WalletAuth:        WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
			StoredCredentials: []string{"http://example.edu/credentials/1877"},
			ProofOptions: &wallet.ProofOptions{
				Controller: "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv464",
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, ProveFromWalletErrorCode, errMsg)
		b.Reset()

		cmdErr = cmd.Issue(&b, getReader(t, &IssueRequest{
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
			Credential: []byte(sampleUDCVC),
			ProofOptions: &wallet.ProofOptions{
				Controller: sampleDIDKey,
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, IssueFromWalletErrorCode, wallet.ErrWalletLocked.Error())
		b.Reset()
	})

	t.Run("issue,prove,verify with invalid request", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Verify(&b, getReader(t, &VerifyRequest{
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
		}))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, "invalid option")
		b.Reset()

		const errMsg = "invalid character"

		cmdErr = cmd.Prove(&b, bytes.NewBufferString("----"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, errMsg)
		b.Reset()

		cmdErr = cmd.Verify(&b, bytes.NewBufferString("----"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, errMsg)
		b.Reset()

		cmdErr = cmd.Issue(&b, bytes.NewBufferString("----"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, errMsg)
		b.Reset()
	})
}

func TestCommand_Derive(t *testing.T) {
	const sampleUser1 = "sample-user-01"

	mockctx := newMockProvider(t)
	mockctx.VDRegistryValue = getMockDIDKeyVDR()

	createSampleUserProfile(t, mockctx, &CreateOrUpdateProfileRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	token, lock := unlockWallet(t, mockctx, &UnlockWalletRequest{
		UserID:             sampleUser1,
		LocalKMSPassphrase: samplePassPhrase,
	})

	defer lock()

	addContent(t, mockctx, &AddContentRequest{
		Content:     []byte(sampleBBSVC),
		ContentType: "credential",
		WalletAuth:  WalletAuth{UserID: sampleUser1, Auth: token},
	})

	// prepare frame
	var frameDoc map[string]interface{}

	require.NoError(t, json.Unmarshal([]byte(sampleFrame), &frameDoc))

	t.Run("derive a credential from stored credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Derive(&b, getReader(t, &DeriveRequest{
			WalletAuth:         WalletAuth{UserID: sampleUser1, Auth: token},
			StoredCredentialID: "http://example.edu/credentials/1872",
			DeriveOptions: &wallet.DeriveOptions{
				Frame: frameDoc,
				Nonce: uuid.New().String(),
			},
		}))
		require.NoError(t, cmdErr)

		var response DeriveResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.NotEmpty(t, response)
		require.NotEmpty(t, response.Credential)
	})

	t.Run("derive a credential from raw credential", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Derive(&b, getReader(t, &DeriveRequest{
			WalletAuth:    WalletAuth{UserID: sampleUser1, Auth: token},
			RawCredential: []byte(sampleBBSVC),
			DeriveOptions: &wallet.DeriveOptions{
				Frame: frameDoc,
				Nonce: uuid.New().String(),
			},
		}))
		require.NoError(t, cmdErr)

		var response DeriveResponse
		require.NoError(t, json.NewDecoder(&b).Decode(&response))
		require.NotEmpty(t, response)
		require.NotEmpty(t, response.Credential)
	})

	t.Run("derive a credential using invalid auth", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Derive(&b, getReader(t, &DeriveRequest{
			WalletAuth:         WalletAuth{UserID: sampleUser1, Auth: sampleFakeTkn},
			StoredCredentialID: "http://example.edu/credentials/1872",
			DeriveOptions: &wallet.DeriveOptions{
				Frame: frameDoc,
				Nonce: uuid.New().String(),
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, DeriveFromWalletErrorCode, "invalid auth token")
		require.Empty(t, b.Bytes())
	})

	t.Run("derive a credential using invalid profile", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Derive(&b, getReader(t, &DeriveRequest{
			WalletAuth:         WalletAuth{UserID: sampleUserID, Auth: sampleFakeTkn},
			StoredCredentialID: "http://example.edu/credentials/1872",
			DeriveOptions: &wallet.DeriveOptions{
				Frame: frameDoc,
				Nonce: uuid.New().String(),
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, DeriveFromWalletErrorCode, "profile does not exist")
		require.Empty(t, b.Bytes())
	})

	t.Run("derive a credential using invalid request", func(t *testing.T) {
		cmd := New(mockctx)

		var b bytes.Buffer

		cmdErr := cmd.Derive(&b, bytes.NewBufferString("--"))
		validateError(t, cmdErr, command.ValidationError, InvalidRequestErrorCode, "invalid character")
		require.Empty(t, b.Bytes())
		b.Reset()

		cmdErr = cmd.Derive(&b, getReader(t, &DeriveRequest{
			WalletAuth: WalletAuth{UserID: sampleUser1, Auth: token},
			DeriveOptions: &wallet.DeriveOptions{
				Frame: frameDoc,
				Nonce: uuid.New().String(),
			},
		}))
		validateError(t, cmdErr, command.ExecuteError, DeriveFromWalletErrorCode, "failed to resolve request")
		require.Empty(t, b.Bytes())
	})
}

func createSampleUserProfile(t *testing.T, ctx *mockprovider.Provider, request *CreateOrUpdateProfileRequest) {
	cmd := New(ctx)
	require.NotNil(t, cmd)

	var l bytes.Buffer
	cmdErr := cmd.CreateProfile(&l, getReader(t, request))
	require.NoError(t, cmdErr)
}

func getReader(t *testing.T, v interface{}) io.Reader {
	vcReqBytes, err := json.Marshal(v)
	require.NoError(t, err)

	return bytes.NewBuffer(vcReqBytes)
}

func getUnlockToken(t *testing.T, b bytes.Buffer) string {
	var response UnlockWalletResponse

	require.NoError(t, json.NewDecoder(&b).Decode(&response))

	return response.Token
}

func unlockWallet(t *testing.T, ctx *mockprovider.Provider, request *UnlockWalletRequest) (string, func()) {
	cmd := New(ctx)

	var b bytes.Buffer

	cmdErr := cmd.Open(&b, getReader(t, &request))
	require.NoError(t, cmdErr)

	return getUnlockToken(t, b), func() {
		cmdErr = cmd.Close(&b, getReader(t, &LockWalletRequest{UserID: request.UserID}))
		if cmdErr != nil {
			t.Log(t, cmdErr)
		}
	}
}

func addContent(t *testing.T, ctx *mockprovider.Provider, request *AddContentRequest) {
	cmd := New(ctx)

	var b bytes.Buffer
	defer b.Reset()

	cmdErr := cmd.Add(&b, getReader(t, &request))
	require.NoError(t, cmdErr)
}

func validateError(t *testing.T, err command.Error,
	expectedType command.Type, expectedCode command.Code, contains string) {
	require.Error(t, err)
	require.Equal(t, err.Type(), expectedType)
	require.Equal(t, err.Code(), expectedCode)

	if contains != "" {
		require.Contains(t, err.Error(), contains)
	}
}

func newMockProvider(t *testing.T) *mockprovider.Provider {
	t.Helper()

	loader, err := jsonldtest.DocumentLoader()
	require.NoError(t, err)

	return &mockprovider.Provider{
		StorageProviderValue:      mockstorage.NewMockStoreProvider(),
		JSONLDDocumentLoaderValue: loader,
	}
}

func getMockDIDKeyVDR() *mockvdr.MockVDRegistry {
	return &mockvdr.MockVDRegistry{
		ResolveFunc: func(didID string, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
			if strings.HasPrefix(didID, "did:key:") {
				k := key.New()

				d, e := k.Read(didID)
				if e != nil {
					return nil, e
				}

				return d, nil
			}

			return nil, fmt.Errorf("did not found")
		},
	}
}

func parseCredential(t *testing.T, b bytes.Buffer) *verifiable.Credential {
	var response struct {
		Credential json.RawMessage
	}

	require.NoError(t, json.NewDecoder(&b).Decode(&response))

	loader, err := jsonldtest.DocumentLoader()
	require.NoError(t, err)

	vc, err := verifiable.ParseCredential(response.Credential, verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(loader))
	require.NoError(t, err)

	return vc
}

func parsePresentation(t *testing.T, b bytes.Buffer) *verifiable.Presentation {
	var response struct {
		Presentation json.RawMessage
	}

	require.NoError(t, json.NewDecoder(&b).Decode(&response))

	loader, err := jsonldtest.DocumentLoader()
	require.NoError(t, err)

	vp, err := verifiable.ParsePresentation(response.Presentation, verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(loader))
	require.NoError(t, err)

	return vp
}
