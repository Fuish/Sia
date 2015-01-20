package consensus

// A block is composed of many transactions. Blocks that have transactions that
// depend on other transactions, but the transactions must all be applied in a
// deterministic order. Transactions cannot have inter-dependencies, meaning
// that an output cannot be created and then spent in the same transaction. As
// far as diffs are concenred, this means that the elements of a transaction
// diff should be able to be applied in any order and the outcome should be the
// same. The elements of a block diff however must be applied in a specific
// order, as transactions may depend on each other.

// An OutputDiff indicates an output that has either been added to or removed
// from the unspent outputs set. New=true means that the output was added when
// the block was applied, and new=false means that the output was deleted when
// the block was applied.
type OutputDiff struct {
	New    bool
	ID     OutputID
	Output Output
}

type ContractDiff struct {
	New        bool
	ContractID ContractID
	Contract   FileContract
}

// A TransactionDiff is the diff that gets applied to the state in the presense
// of a transaction.
type TransactionDiff struct {
	OutputDiffs   []OutputDiff
	ContractDiffs []ContractDiff
}

// A BlockDiff contains the list of changes that happened to the state when
// changing from one block to another. A diff is bi-directional, and
// deterministically applied.
type BlockDiff struct {
	CatalystBlock    BlockID // Which block was used to derive the diffs.
	TransactionDiffs []TransactionDiff
	BlockChanges     TransactionDiff // Changes specific to the block being in place - subsidies and contract maintenance.
}

// A ConsensusChange is a list of block diffs that have been applied to the
// state. The ConsensusChange is sent to everyone who has subscribed to the
// state.
type ConsensusChange struct {
	InvertedBlocks []BlockDiff
	AppliedBlocks  []BlockDiff
}

// notifySubscribers sends a ConsensusChange notification to every subscriber
//
// The sending is done in a separate goroutine to prevent deadlock if one
// subscriber becomes unresponsive.
//
// TODO: What happens if a subscriber stops pulling info from their channel. If
// they don't close the channel but stop pulling out elements, will the system
// lock up? If something stops responding suddenly, there needs to be a way to
// keep going, the state just deletes, closes, or ignores the channel or
// something. Perhaps the state will close the channel if the buffer fills up,
// assuming that the component has shut down unexpectedly. If the component was
// just being slow, it can do some catching up and re-subscribe. If we do end
// up closing subscription channels then we should switch from a slice to a
// map for s.consensusSubscriptions.
func (s *State) notifySubscribers(cc ConsensusChange) {
	for _, sub := range s.consensusSubscriptions {
		sub <- cc
	}

	for _, sub := range s.subscriptions {
		select {
		case sub <- struct{}{}:
			// Receiver has been notified of an update.
		default:
			// Receiver already has notification to check for updates.
		}
	}
}

// ConsensusSubscribe returns a channel that will receive a ConsensusChange
// notification each time that the consensus changes (from incoming blocks or
// invalidated blocks, etc.).
//
// TODO: Depricate in favor not sending a whole diff down the channel.
func (s *State) ConsensusSubscribe() (alert chan ConsensusChange) {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert = make(chan ConsensusChange)
	s.consensusSubscriptions = append(s.consensusSubscriptions, alert)
	return
}

// Subscribe allows a module to subscribe to the state, which means that it'll
// receive a notification (in the form of an empty struct) each time the state
// gets a new block.
func (s *State) Subscribe() (alert chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	alert = make(chan struct{})
	s.subscriptions = append(s.subscriptions, alert)
	return
}
