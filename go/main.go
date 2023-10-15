package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type key struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type keys struct {
	Kvs []*key `json:"kvs"`
}

func main() {
	// curl()
	// httpsclient()
	etcdClient()
}

func curl() {
	keys := &keys{}

	cmd := exec.Command("curl", "--insecure", "--cert", "peer.crt", "--key", "peer.key", "-X", "POST", "https://192.168.49.2:2379/v3/kv/range", "-d", `{"key": "AA==", "range_end": "AA=="}`)

	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}

	json.Unmarshal(out, keys)

	for _, k := range keys.Kvs {
		kbytes, _ := base64.StdEncoding.DecodeString(k.Key)
		vbytes, _ := base64.StdEncoding.DecodeString(k.Value)
		k.Key = string(kbytes)
		k.Value = string(vbytes)
		fmt.Printf("key: %v\n", k.Key)
		fmt.Printf("value: %v\n", k.Value)
	}
}

func httpsclient() {
	kp, _ := tls.LoadX509KeyPair("peer.crt", "peer.key")

	certs := []tls.Certificate{kp}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, Certificates: certs},
	}

	client := &http.Client{Transport: tr}

	data := []byte(`{"key": "AA==", "range_end": "AA=="}`)

	bodyReader := bytes.NewReader(data)

	resp, err := client.Post("https://192.168.49.2:2379/v3/kv/range", "application/json", bodyReader)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Printf("resp: %v\n", resp.Body)

	keys := &keys{}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	json.Unmarshal(out, keys)

	for _, k := range keys.Kvs {
		kbytes, _ := base64.StdEncoding.DecodeString(k.Key)
		vbytes, _ := base64.StdEncoding.DecodeString(k.Value)
		k.Key = string(kbytes)
		k.Value = string(vbytes)
		fmt.Printf("key: %v\n", k.Key)
		fmt.Printf("value: %v\n", k.Value)
	}
}

func etcdClient() {
	kp, _ := tls.LoadX509KeyPair("peer.crt", "peer.key")

	certs := []tls.Certificate{kp}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.58.2:2379"},
		DialTimeout: 5 * time.Second,
		TLS:         &tls.Config{InsecureSkipVerify: true, Certificates: certs},
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	resp, _ := cli.KV.Get(context.Background(), "", clientv3.WithPrefix())

	var values []byte
	var keys string

	for _, k := range resp.Kvs {
		fmt.Printf("key: %s\n", k.Key)
		keys = keys + fmt.Sprintf("%s\n", k.Key)
		// fmt.Printf("value: %s\n", k.Value)
		values = append(values, k.Value...)
	}

	os.WriteFile("val2", []byte(keys), 0644)
}
