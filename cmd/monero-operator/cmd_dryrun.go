package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cirocosta/monero-operator/pkg/apis/utxo.com.br/v1alpha1"
	"github.com/cirocosta/monero-operator/pkg/reconciler"
)

type DryRunCommand struct {
	File string `long:"file" short:"f" required:"true" description:"manifest with MoneroNodeSet definition"`
}

func (c *DryRunCommand) Execute(_ []string) error {
	nodeSet, err := c.MoneroNodeSetFromFile(c.File)
	if err != nil {
		return fmt.Errorf("monero nodeset from file '%s': %w", c.File, err)
	}

	nodeSet.ApplyDefaults()

	rec := &reconciler.MoneroNodeSetReconciler{Log: log.Log}
	objs, err := rec.GenerateObjects(context.TODO(), nodeSet)
	if err != nil {
		return fmt.Errorf("generate objects: %w", err)
	}

	if err := c.WriteObjects(os.Stdout, objs); err != nil {
		return fmt.Errorf("write objs: %w", err)
	}

	return nil
}

func (c *DryRunCommand) WriteObjects(w io.Writer, objs []client.Object) error {
	const YAMLDocumentSeparator = "---\n"

	encoder := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	for _, o := range objs {
		w.Write([]byte(YAMLDocumentSeparator))
		if err := encoder.Encode(o, w); err != nil {
			return fmt.Errorf("marhsal: %w", err)
		}
	}

	return nil
}

func (c *DryRunCommand) MoneroNodeSetFromFile(fname string) (*v1alpha1.MoneroNodeSet, error) {
	res := &v1alpha1.MoneroNodeSet{}

	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open '%s': %w", fname, err)
	}
	defer f.Close()

	if err := kyaml.NewYAMLOrJSONDecoder(f, 1<<10).Decode(res); err != nil {
		return nil, fmt.Errorf("decode into monero nodeset: %w", err)
	}

	return res, nil
}

func init() {
	parser.AddCommand("dry-run",
		"Perform a dry-run of the reconciler",
		"Run the reconciler without actually submitting anything",
		&DryRunCommand{},
	)
}
