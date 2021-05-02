package reconciler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MonerodConfig map[string]string

func (c *MonerodConfig) Conf() string {
	var (
		buffer = new(bytes.Buffer)
		mmap   = *c
	)

	keys := reflect.ValueOf(mmap).MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Interface().(string) < keys[j].Interface().(string)
	})

	for _, k := range keys {
		v := mmap[k.Interface().(string)]
		buffer.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	return buffer.String()
}

func (c *MonerodConfig) Digest() string {
	conf := c.Conf()

	h := sha256.New()
	h.Write([]byte(conf))

	return hex.EncodeToString(h.Sum(nil))
}

func NewDefaultMonerodConfig() MonerodConfig {
	return MonerodConfig{
		"data-dir": MonerodDataVolumeMountPath,

		"enable-dns-blocklist":      "1",
		"enforce-dns-checkpointing": "1",
		"confirm-external-bind":     "1",

		"in-peers":  "1024",
		"out-peers": "1024",

		"limit-rate-down": "1048576",
		"limit-rate-up":   "1048576",

		"log-file":          "/var/log/monero/monerod.log",
		"max-log-file-size": "0",

		"no-igd": "1",

		"p2p-bind-ip":              "0.0.0.0",
		"p2p-bind-port":            "18080",
		"rpc-bind-ip":              "0.0.0.0",
		"rpc-bind-port":            "18081",
		"rpc-restricted-bind-ip":   "0.0.0.0",
		"rpc-restricted-bind-port": "18089",
	}
}

func (c *MonerodConfig) ConfigMap(name, namespace string) *corev1.ConfigMap {
	obj := &corev1.ConfigMap{}

	obj.TypeMeta = metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: corev1.SchemeGroupVersion.Identifier(),
	}

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	obj.Data = map[string]string{
		"monerod.conf": c.Conf(),
	}

	return obj
}
