<!--
order: 3
-->

# Use Cases

Check out the 2 use cases for the Haqq project. {synopsis}

## HAQQ chain

The Haqq blockchain provides Ethereum developers to deploy their smart contracts to the
Haqq EVM and get the benefits of a fast-finality Proof-of-Stake (PoS) chain. Developers will
also benefit from highly-reliable clients from testnets can be used to test and deploy their
contracts.

Haqq will also offer built-in interoperability functionalities with other Cosmos and BFT chains by using [IBC](https://ibcprotocol.org/). Developers can also benefit from using a bridge network to enable interoperability between mainnet Ethereum and Haqq.

## EVM module dependency

The EVM module (aka [x/evm](https://github.com/haqq-network/haqq/tree/master/x/evm)) packaged inside
Haqq can be used separately as its own standalone module. This can be added as a dependency to
any Cosmos chain, which will allow for smart contract support.

Importing EVM module can also enable use cases such as Proof-of-Authority
([PoA](https://en.wikipedia.org/wiki/Proof_of_authority)) chains for enterprise and consortium
projects. Every chain on Cosmos is an [application-specific
blockchain](https://docs.cosmos.network/master/intro/why-app-specific.html) that is customized for
the business logic defined by a single application. Thus, by using a predefined validator set and
the EVM module as a dependency, enables projects with fast finality, interoperability as well as
Proof-of-Stake (PoS) consensus.

## Trade offs

Either option above will allow for fast finality, using a PoS consensus engine. Using the EVM module
as a dependency will require the importing of the EVM and the maintaining of the chain (including
validator sets, code upgrades/conformance, community engagement, incentives, etc), thus it incurs on a
higher operation cost. The benefit of importing the EVM module to your chains is that it allows for
granular control over the network and chain specific configurations/features that may not be
available in the Haqq chain such as developing a module or importing a third-party one.

Using Haqq chain will allow for the direct deployment of smart contracts to the Haqq
network. Utilizing the Haqq client will defer the chain maintenance to the Haqq network
and allow for the participation in a more mature blockchain. The Haqq client will also offer
(in the near future) IBC compatibility which allows for interoperability between different network.

|                                         | Haqq Chain     | x/evm dependency |
|-----------------------------------------|-----------------|------------------|
| Maintenance                             | Lower           | Higher           |
| Sovereignty (validator, config, params) | Lower           | Higher           |
| Storage requirements                    | Lower           | Higher           |

## Next {hide}

Read the available Haqq [resources](./resources.md) {hide}
