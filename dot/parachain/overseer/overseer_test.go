package overseer

import (
	"fmt"
	parachainTypes "github.com/ChainSafe/gossamer/dot/parachain/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type ExampleSubsystem1 struct {
	name string
}

func (e *ExampleSubsystem1) Run(context *Context) error {
	fmt.Printf("Run %v\n", e.name)
	err := e.initialize(*context)
	if err != nil {
		return fmt.Errorf("initialize %v: %w", e.name, err)
	}
	return nil
}

func (e *ExampleSubsystem1) ProcessActiveLeavesUpdate(update ActiveLeavesUpdate) error {
	fmt.Printf("ParticipationHandler received active leaves update %v\n", update)
	return nil
}

func (e *ExampleSubsystem1) waitForFirstLeaf(context Context) (*ActivatedLeaf, error) {
	for {
		select {
		case overseerSignal := <-context.Receiver:
			return overseerSignal.(*ActivatedLeaf), nil
		}
	}
}

func (e *ExampleSubsystem1) initialize(context Context) error {
	firstLeaf, err := e.waitForFirstLeaf(context)
	if err != nil {
		return fmt.Errorf("initialize %v: %w", e.name, err)
	}

	return e.handleStartup(context, firstLeaf)
}

func (e *ExampleSubsystem1) handleStartup(context Context, initalHead *ActivatedLeaf) error {
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("%v doing %v\n", e.name, initalHead)
			context.Sender.SendMessage(fmt.Sprintf("hello from %v", e.name))
		}
	}()
	return nil
}

func TestStartSubsystems(t *testing.T) {
	overseer := NewOverseer()

	ss1 := &ExampleSubsystem1{
		name: "subSystem 1",
	}
	ss2 := &ExampleSubsystem1{
		name: "subSystem 2",
	}
	overseer.RegisterSubSystem(ss1)
	overseer.RegisterSubSystem(ss2)
	overseer.Start()
	time.Sleep(time.Millisecond * 500)
	err := overseer.sendActiveLeaf(parachainTypes.BlockNumber(11))
	require.NoError(t, err)

	time.Sleep(5 * time.Second)
	overseer.stop()
}
