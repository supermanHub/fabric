/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"time"

	"github.com/hyperledger/fabric/common/policies"
	"github.com/hyperledger/fabric/msp"
	cb "github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ChannelConfig stores the common channel config
type ChannelConfig interface {
	// HashingAlgorithm returns the default algorithm to be used when hashing
	// such as computing block hashes, and CreationPolicy digests
	HashingAlgorithm() func(input []byte) []byte

	// BlockDataHashingStructureWidth returns the width to use when constructing the
	// Merkle tree to compute the BlockData hash
	BlockDataHashingStructureWidth() uint32

	// OrdererAddresses returns the list of valid orderer addresses to connect to to invoke Broadcast/Deliver
	OrdererAddresses() []string
}

type OrgConfig interface {
	// Name returns the name this org is referred to in config
	Name() string

	// MSPID returns the MSP ID associated with this org
	MSPID() string
}

// ApplicationOrgConfig stores the per org application config
type ApplicationOrgConfig interface {
	OrgConfig

	// AnchorPeers returns the list of gossip anchor peers
	AnchorPeers() []*pb.AnchorPeer
}

// ApplicationConfig stores the common shared application config
type ApplicationConfig interface {
	// Organizations returns a map of org ID to ApplicationOrgConfig
	Organizations() map[string]ApplicationOrgConfig
}

// OrdererConfig stores the common shared orderer config
type OrdererConfig interface {
	// ConsensusType returns the configured consensus type
	ConsensusType() string

	// BatchSize returns the maximum number of messages to include in a block
	BatchSize() *ab.BatchSize

	// BatchTimeout returns the amount of time to wait before creating a batch
	BatchTimeout() time.Duration

	// ChainCreationPolicyNames returns the policy names which are allowed for chain creation
	// This field is only set for the system ordering chain
	ChainCreationPolicyNames() []string

	// KafkaBrokers returns the addresses (IP:port notation) of a set of "bootstrap"
	// Kafka brokers, i.e. this is not necessarily the entire set of Kafka brokers
	// used for ordering
	KafkaBrokers() []string

	// IngressPolicyNames returns the name of the policy to validate incoming broadcast messages against
	IngressPolicyNames() []string

	// EgressPolicyNames returns the name of the policy to validate incoming broadcast messages against
	EgressPolicyNames() []string
}

// Handler provides a hook which allows other pieces of code to participate in config proposals
// TODO, this should probably be renamed to ValueHandler
type Handler interface {
	Transactional

	// ProposeConfig called when config is added to a proposal
	ProposeConfig(key string, configValue *cb.ConfigValue) error
}

// Manager provides a mechanism to query and update config
type Manager interface {
	Resources

	// Apply attempts to apply a configtx to become the new config
	Apply(configtx *cb.Envelope) error

	// Validate attempts to validate a new configtx against the current config state
	Validate(configtx *cb.Envelope) error

	// ConfigEnvelope returns the *cb.ConfigEnvelope from the last successful Apply
	ConfigEnvelope() *cb.ConfigEnvelope

	// ChainID retrieves the chain ID associated with this manager
	ChainID() string

	// Sequence returns the current sequence number of the config
	Sequence() uint64
}

// Resources is the common set of config resources for all channels
// Depending on whether chain is used at the orderer or at the peer, other
// config resources may be available
type Resources interface {
	// PolicyManager returns the policies.Manager for the channel
	PolicyManager() policies.Manager

	// ChannelConfig returns the ChannelConfig for the chain
	ChannelConfig() ChannelConfig

	// OrdererConfig returns the configtxorderer.SharedConfig for the channel
	OrdererConfig() OrdererConfig

	// ApplicationConfig returns the configtxapplication.SharedConfig for the channel
	ApplicationConfig() ApplicationConfig

	// MSPManager returns the msp.MSPManager for the chain
	MSPManager() msp.MSPManager
}

// Transactional is an interface which allows for an update to be proposed and rolled back
type Transactional interface {
	// BeginConfig called when a config proposal is begun
	BeginConfig(groups []string) ([]Handler, error)

	// RollbackConfig called when a config proposal is abandoned
	RollbackConfig()

	// CommitConfig called when a config proposal is committed
	CommitConfig()
}

// PolicyHandler is used for config updates to policy
type PolicyHandler interface {
	Transactional

	ProposePolicy(key string, path []string, policy *cb.ConfigPolicy) error
}

// Initializer is used as indirection between Manager and Handler to allow
// for single Handlers to handle multiple paths
type Initializer interface {
	Handler

	Resources

	// PolicyProposer returns the PolicyHandler to handle updates to policy
	PolicyHandler() PolicyHandler
}
