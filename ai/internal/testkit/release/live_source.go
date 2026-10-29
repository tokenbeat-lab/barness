package release

import (
	"flag"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"io"
	"path/filepath"
)

const executionSourceInput = "@execution-source/"

// A supplement is an explicit historical input, not a change to the bundle.
// Its canonical map must still match the original producer grant's source pin.
func liveInputs(root, dir, supplement string) (map[string]string, string, error) {
	if supportmatrix.IsXAIImagesCapture(dir) {
		if supplement != "" {
			return nil, "", fmt.Errorf("Imagine source override forbidden")
		}
		inputs, e := supportmatrix.XAIImagesInputs(dir)
		return inputs, "", e
	}
	if supportmatrix.IsMiniMaxThinkingComposition(dir) {
		if supplement != "" {
			return nil, "", fmt.Errorf("MiniMax composition source override forbidden")
		}
		inputs, e := supportmatrix.MiniMaxThinkingCompositionInputs(dir)
		return inputs, "", e
	}
	if supportmatrix.IsKimiComposition(dir) {
		if supplement != "" {
			return nil, "", fmt.Errorf("Kimi composition does not accept a generic source supplement")
		}
		inputs, err := supportmatrix.KimiCompositionInputs(dir)
		return inputs, "", err
	}
	inputs, err := liveInventory(dir)
	if err != nil || supplement == "" {
		return inputs, "", err
	}
	path, err := recordPath(root, supplement)
	if err != nil {
		return nil, "", err
	}
	hash, err := recordHash(path)
	if err != nil {
		return nil, "", err
	}
	if _, exists := inputs[executionSourceInput+supplement]; exists {
		return nil, "", fmt.Errorf("source input collides with bundle inventory")
	}
	inputs[executionSourceInput+supplement] = hash
	return inputs, path, nil
}

type liveInvocation struct{ bundle, out, supplement string }

func readLiveConsumer(root string, ref FileRef) (LiveConsumerReceipt, liveInvocation, error) {
	var proof LiveConsumerReceipt
	var args liveInvocation
	path, err := verifyRecordFile(root, ref)
	if err != nil {
		return proof, args, err
	}
	if recordJSON(path, &proof) != nil {
		return proof, args, fmt.Errorf("invalid native consumer receipt")
	}
	flags := flag.NewFlagSet("native-live-consumer", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&args.bundle, "live-bundle", "", "")
	flags.StringVar(&args.out, "out", "", "")
	flags.StringVar(&args.supplement, "execution-source", "", "")
	if len(proof.Argv) == 0 || filepath.Base(proof.Argv[0]) != "releasegate" || flags.Parse(proof.Argv[1:]) != nil || flags.NArg() != 0 {
		return proof, args, fmt.Errorf("native consumer invocation invalid")
	}
	return proof, args, nil
}
